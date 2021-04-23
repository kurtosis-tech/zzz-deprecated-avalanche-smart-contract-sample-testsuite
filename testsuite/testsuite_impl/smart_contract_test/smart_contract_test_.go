package smart_contract_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/smart_contracts/bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/mieubrisse/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/mieubrisse/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/topology"
	"github.com/mieubrisse/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/mieubrisse/avalanchego-kurtosis/kurtosis/avalanche/tests/testconstants"
	"github.com/mieubrisse/avalanchego-kurtosis/kurtosis/kurtosis/networksavalanche"
	"github.com/palantir/stacktrace"
	"github.com/status-im/keycard-go/hexutils"
	"time"
)

const (
	contractBinaryHex = "608060405234801561001057600080fd5b5061012f806100206000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c806360fe47b11460375780636d4ce63c14604f575b600080fd5b604d600480360381019060499190608f565b6069565b005b60556073565b6040516060919060c2565b60405180910390f35b8060008190555050565b60008054905090565b60008135905060898160e5565b92915050565b60006020828403121560a057600080fd5b600060ac84828501607c565b91505092915050565b60bc8160db565b82525050565b600060208201905060d5600083018460b5565b92"

	bootstrapNodeServiceIdPrefix = "bootstrapNode-"
)

var definedNetwork =

type SmartContractTest struct {
	avalancheImage string
	networkDefinition *networkbuilder.Network
}

func (test SmartContractTest) GetTestConfiguration() testsuite.TestConfiguration {
	return testsuite.TestConfiguration{
		IsPartitioningEnabled: false,
		FilesArtifactUrls:     map[services.FilesArtifactID]string{},
	}
}

func (test SmartContractTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	networkConfiguration := networkbuilder.New().
		Image(test.avalancheImage).
		SnowSize(3, 3)

	var i uint = 1
	for _, staker := range constants.DefaultLocalNetGenesisConfig.Stakers {
		nodeConfig := networkbuilder.NewNode(getBootstrapNodeId(i)).
			Image(test.avalancheImage).
			IsStaking(true).
			BootstrapNode(true).
			BootstrapNodeID(int(i)).
			ConnectedBTNodeIDs(networkConfiguration.GetConnectedBTNodeIDs()).
			PrivateKey(staker.PrivateKey)
		nodeConfig.TLSCert(staker.TLSCert)
		networkConfiguration.AddNode(nodeConfig)
		networkConfiguration.ConnectedBTNodeIDs(staker.NodeID)
		i++
	}
	networkConfiguration.HasBootstrapNodes(true)

	avalancheNetwork := networksavalanche.NewAvalancheNetwork(networkCtx, test.avalancheImage)

	// first setup bootstrap nodes
	var nodeChecker []*services.DefaultAvailabilityChecker
	for i := 1; i <= networkConfiguration.GetNumBootstrapNodes(); i++ {
		if bootstrapNode, ok := networkConfiguration.Nodes[fmt.Sprintf("bootstrapNode-%d", i)]; ok {
			if bootstrapNode.IsBootstrapNode() {
				_, checker, err := avalancheNetwork.CreateNodeNoCheck(networkConfiguration, bootstrapNode)
				if err != nil {
					return nil, stacktrace.Propagate(err, "An error occurred creating a new Node")
				}
				nodeChecker = append(nodeChecker, checker)
			}
		}
	}

	for _, checker := range nodeChecker {
		err := checker.WaitForStartup(15*time.Second, 10)
		if err != nil {
			panic(err)
		}
	}

	for _, node := range networkConfiguration.Nodes {
		if !node.IsBootstrapNode() {
			_, err := avalancheNetwork.CreateNode(networkConfiguration, node)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating a new Node")
			}
		}
	}

	test.networkDefinition = networkConfiguration
	return networkConfiguration, nil
}

func (test SmartContractTest) Run(network networks.Network, testCtx testsuite.TestContext) {
	topology := topology.New(network, &testCtx).LoadDefinedNetwork(test.networkDefinition)

	node1Id := getBootstrapNodeId(1)
	node1 := topology.Node(node1Id)
	node1XAddr := node1.XAddress
	node1AvalancheGoClient := node1.GetClient()

	// 1000 AVAX
	topology.Genesis().FundXChainAddresses([]string{node1XAddr}, 1000000000000)

	cChainAddrStr, err := node1AvalancheGoClient.CChainAPI().CreateAddress(node1.UserPass)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred creating a C-Chain address for node: %v", node1Id))
	}

	cChainAddrBytes := common.HexToAddress(cChainAddrStr)
	topology.Genesis().MoveBalanceToCChain(cChainAddrBytes, testconstants.TxFee)

	// topology.Genesis().MoveBalanceToCChain()


	/*
	ethClient := node1.GetClient().CChainEthAPI()

	ethClient

	bindings.DeployStorage(, ethClient)
	ethClient.c
	 */

}

func (test SmartContractTest) GetSetupTimeout() time.Duration {
	return 60 * time.Second
}

func (test SmartContractTest) GetExecutionTimeout() time.Duration {
	return 60 * time.Second
}

func getBootstrapNodeId(idx uint) string {
	return fmt.Sprintf("bootstrapNode-%d", idx)
}
