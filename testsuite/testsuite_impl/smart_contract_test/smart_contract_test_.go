package smart_contract_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/smart_contracts/bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/topology"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/tests/testconstants"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/kurtosis/networksavalanche"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math/big"
	"time"
)

const (
	// TODO Get this from the node config somehow, rather than hardcoding it
	nodeHttpPort = 9650
)

type SmartContractTest struct {
	avalancheImage string
	// TODO make this not a mutable variable that gets set in Setup
	networkDefinition *networkbuilder.Network
}

func NewSmartContractTest(avalancheImage string) *SmartContractTest {
	return &SmartContractTest{avalancheImage: avalancheImage}
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

	node1CChainApi := node1AvalancheGoClient.CChainAPI()

	cChainAddrStr, err := node1CChainApi.CreateAddress(node1.UserPass)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred creating a C-Chain address for node: %v", node1Id))
	}

	cChainAddrBytes := common.HexToAddress(cChainAddrStr)
	topology.Genesis().MoveBalanceToCChain(cChainAddrBytes, testconstants.TxFee)

	_, privKeyHex, err := node1CChainApi.ExportKey(node1.UserPass, cChainAddrStr)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred exporting the private key for the C-Chain address '%v'", cChainAddrStr))
	}

	privKeyEcdsa, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred converting the C-Chain private key hex '%v' to an ECDSA key", privKeyHex))
	}

	transactor := bind.NewKeyedTransactor(privKeyEcdsa)

	uri := fmt.Sprintf("ws://%s:%d/ext/bc/C/ws", node1.GetIPAddress(), nodeHttpPort)
	node1EthClient, err := ethclient.Dial(uri)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting an ethclient for URI '%v'", uri))
	}
	_, _, contract, err := bindings.DeployStorage(transactor, node1EthClient)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred deploying the contract on the C-Chain"))
	}

	var valueToStore int64 = 20
	// TODO wait for transaction to get accepted
	_, err = contract.Store(transactor, big.NewInt(valueToStore))
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred storing value '%v' in the contract", valueToStore))
	}

	retrievedValue, err := contract.Retrieve(nil)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred retrieving the value stored in the contract"))
	}

	logrus.Infof("Retrieved value: %v", retrievedValue)
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
