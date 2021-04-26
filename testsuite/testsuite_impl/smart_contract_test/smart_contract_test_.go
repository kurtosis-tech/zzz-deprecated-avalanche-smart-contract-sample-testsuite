package smart_contract_test

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/networkbuilder"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/builder/topology"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/avalanche/tests/testconstants"
	"github.com/ava-labs/avalanchego-kurtosis/kurtosis/kurtosis/networksavalanche"
	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/smart_contracts/bindings"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math/big"
	"strings"
	"time"
)

const (
	// TODO Get this from the node config somehow, rather than hardcoding it
	nodeHttpPort = 9650

	hexStrIndicatorLeader = "0x"

	maxNumCheckTransactionMinedRetries = 10
	timeBetweenCheckTransactionMinedRetries = 1 * time.Second
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

func (test *SmartContractTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
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
	return avalancheNetwork, nil
}

func (test SmartContractTest) Run(network networks.Network, testCtx testsuite.TestContext) {
	runningNetwork := topology.New(network, &testCtx)

	firstValidatorId := ""
	for _, node := range test.networkDefinition.Nodes {
		runningNetwork.AddNode(node.ID, node.ID, constants.DefaultPassword)
		if firstValidatorId == "" {
			firstValidatorId = node.ID
		}
	}
	runningNetwork.AddGenesis(
		firstValidatorId,
		testconstants.GenesisUsername,
		testconstants.GenesisPassword,
	)

	node1Id := getBootstrapNodeId(1)
	node1 := runningNetwork.Node(node1Id)
	node1AvalancheGoClient := node1.GetClient()

	logrus.Info("Creating C-Chain address...")
	node1CChainApi := node1AvalancheGoClient.CChainAPI()
	// Client ...
	type Client struct {
		requester rpc.EndpointRequester
	}

	xChainPrivateKey, err := node1AvalancheGoClient.XChainAPI().ExportKey(node1.UserPass, node1.XAddress)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred exporting the node's private key from the X-Chain"))
	}
	// TODO Debugging
	logrus.Infof("X-Chain private key: %v", xChainPrivateKey)
	cChainAddrStr, err := node1AvalancheGoClient.CChainAPI().ImportKey(node1.UserPass, xChainPrivateKey)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred importing the node's private key to the C-Chain"))
	}
	/*
	endpointRequester := rpc.NewEndpointRequester(
		fmt.Sprintf("http://%s:%d", node1.GetIPAddress(), nodeHttpPort),
		fmt.Sprintf("/ext/bc/C/avax"),
		"evm",
		10 * time.Second,
	)
	res := &api.JSONAddress{}
	if err := endpointRequester.SendRequest("createAddress", &node1.UserPass, res); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred creating the C-Chain address"))
	}
	cChainAddrStr := res.Address
	 */
	/*
	cChainAddrStr, err := node1CChainApi.CreateAddress(node1.UserPass)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred creating a C-Chain address for node: %v", node1Id))
	}
	 */
	logrus.Infof("C-Chain address '%v' created", cChainAddrStr)

	logrus.Info("Transferring balance to C-Chain address...")
	cChainAddrBytes := common.HexToAddress(cChainAddrStr)
	genesis := runningNetwork.Genesis()
	genesis.MoveBalanceToCChain(cChainAddrBytes, testconstants.TxFee)
	logrus.Info("Balance transferred to C-Chain address")

	logrus.Info("Creating keyed transactor...")
	_, privKeyHexWithLead0x, err := node1CChainApi.ExportKey(node1.UserPass, cChainAddrStr)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred exporting the private key for the C-Chain address '%v'", cChainAddrStr))
	}
	privKeyHex := strings.Replace(privKeyHexWithLead0x, hexStrIndicatorLeader, "", 1)
	logrus.Infof("C-Chain private key hex: %v", privKeyHex)
	privKeyEcdsa, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred converting the C-Chain private key hex '%v' to an ECDSA key", privKeyHex))
	}
	transactor := bind.NewKeyedTransactor(privKeyEcdsa)
	logrus.Info("Keyed transactor created")


	logrus.Info("Creating Geth client...")
	uri := fmt.Sprintf("ws://%s:%d/ext/bc/C/ws", node1.GetIPAddress(), nodeHttpPort)
	node1EthClient, err := ethclient.Dial(uri)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting an ethclient for URI '%v'", uri))
	}
	logrus.Info("Geth client created")

	logrus.Info("Deploying contract...")
	_, contractDeploymentTxn, contract, err := bindings.DeployStorage(transactor, node1EthClient)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred deploying the contract on the C-Chain"))
	}
	if err := waitUntilTransactionMined(node1EthClient, contractDeploymentTxn.Hash()); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred waiting for the contract deployment transaction to be mined"))
	}
	logrus.Info("Contract deployed")

	valueToStore := big.NewInt(20)
	logrus.Infof("Storing value '%v'...", valueToStore)
	// TODO wait for transaction to get accepted
	storeValueTxn, err := contract.Store(transactor, valueToStore)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred storing value '%v' in the contract", valueToStore))
	}
	if err := waitUntilTransactionMined(node1EthClient, storeValueTxn.Hash()); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred waiting for the value-storing transaction to be mined"))
	}
	logrus.Info("Value stored")

	logrus.Info("Retrieving value from contract...")
	retrievedValue, err := contract.Retrieve(&bind.CallOpts{})
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred retrieving the value stored in the contract"))
	}
	logrus.Infof("Retrieved value: %v", retrievedValue)

	testCtx.AssertTrue(valueToStore == retrievedValue, stacktrace.NewError("Retrieved value '%v' != stored value '%v'", retrievedValue, valueToStore))
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

// If we try to use a contract immediately after submission without waiting for it to be mined, we'll get a "no contract code at address" error:
// https://github.com/ethereum/go-ethereum/issues/15930#issuecomment-532144875
func waitUntilTransactionMined(validatorClient *ethclient.Client, transactionHash common.Hash) error {
	for i := 0; i < maxNumCheckTransactionMinedRetries; i++ {
		receipt, err := validatorClient.TransactionReceipt(context.Background(), transactionHash)
		// TODO DEBUGGING
		logrus.Infof("Receipt: %+v", receipt)
		if err == nil && receipt != nil && receipt.BlockNumber != nil {
			return nil
		}
		if i < maxNumCheckTransactionMinedRetries - 1 {
			time.Sleep(timeBetweenCheckTransactionMinedRetries)
		}
	}
	return stacktrace.NewError(
		"Transaction with hash '%v' wasn't mined even after checking %v times with %v between checks",
		transactionHash.Hex(),
		maxNumCheckTransactionMinedRetries,
		timeBetweenCheckTransactionMinedRetries)
}
