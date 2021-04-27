package networks_impl

import (
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
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	// TODO Get this from the node config somehow, rather than hardcoding it
	nodeHttpPort = 9650

	// NOTE: This has to be 1-indexed because NodeInitializer requires it, rather than 0-indexed
	initialBootstrapperIdIdx = 1

	hexStrIndicatorLeader = "0x"

	timeBetweenNodeStartupPolls = 5 * time.Second
	maxNumNodeStartupPolls = 30
)

type SmartContractAvalancheNetwork struct {
	avalancheImage string

	avalancheNetwork *networksavalanche.AvalancheNetwork

	networkConfiguration *networkbuilder.Network

	transactor *bind.TransactOpts
	gethClient *ethclient.Client
}

func NewSmartContractAvalancheNetwork(avalancheImage string, networkCtx *networks.NetworkContext) *SmartContractAvalancheNetwork {
	networkConfiguration := networkbuilder.New().
		Image(avalancheImage).
		SnowSize(3, 3)

	i := initialBootstrapperIdIdx
	for _, staker := range constants.DefaultLocalNetGenesisConfig.Stakers {
		nodeConfig := networkbuilder.NewNode(getBootstrapNodeId(i)).
			Image(avalancheImage).
			IsStaking(true).
			BootstrapNode(true).
			BootstrapNodeID(i).
			ConnectedBTNodeIDs(networkConfiguration.GetConnectedBTNodeIDs()).
			PrivateKey(staker.PrivateKey)
		nodeConfig.TLSCert(staker.TLSCert)
		networkConfiguration.AddNode(nodeConfig)
		networkConfiguration.ConnectedBTNodeIDs(staker.NodeID)
		i++
	}
	networkConfiguration.HasBootstrapNodes(true)

	result := &SmartContractAvalancheNetwork{
		avalancheImage:                 avalancheImage,
		avalancheNetwork: networksavalanche.NewAvalancheNetwork(networkCtx, avalancheImage),
		networkConfiguration:           networkConfiguration,
		transactor: nil,
		gethClient: nil,
	}
	return result
}

// Prepares an Avalanche network for smart contract deployment by starting it, creating a C-Chain address, funding it, etc.
// This function is expected to be used in the Test.Setup phase, with GetTransactor and GetGethClient used in Test.Run phase
func (network *SmartContractAvalancheNetwork) SetupAvalancheNetwork() error {
	if network.transactor != nil || network.gethClient != nil {
		return stacktrace.NewError("Avalanche network already started")
	}

	logrus.Info("Launching bootstrap nodes...")
	bootstrapNodeCheckers := map[string]*services.DefaultAvailabilityChecker{}
	// TODO We have to do this "start from 1" indexing because the NodeInitializer currently requires a specific ServiceID pattern,
	//  instantiated in a particular order (see https://github.com/ava-labs/avalanchego-kurtosis/issues/6 )
	for i := 1; i <= len(constants.DefaultLocalNetGenesisConfig.Stakers); i++ {
		id := getBootstrapNodeId(i)
		nodeConfig, found := network.networkConfiguration.Nodes[id]
		if !found {
			return stacktrace.NewError("Expected a node config for ID '%v', but none was found", id)
		}
		_, checker, err := network.avalancheNetwork.CreateNodeNoCheck(network.networkConfiguration, nodeConfig)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating bootstrapper node with ID '%v'", id)
		}
		bootstrapNodeCheckers[id] = checker
	}
	logrus.Info("Bootstrap nodes launched")

	logrus.Info("Waiting for bootstrap nodes to become available...")
	for id, checker := range bootstrapNodeCheckers {
		if err := checker.WaitForStartup(timeBetweenNodeStartupPolls, maxNumNodeStartupPolls); err != nil {
			return stacktrace.Propagate(err, "An error occurred waiting for bootstrapper node '%v' to become available", id)
		}
	}
	logrus.Info("Bootstrap nodes available")

	logrus.Info("Launching non-bootstrap nodes...")
	nonBootstrapNodeCheckers := map[string]*services.DefaultAvailabilityChecker{}
	for id, nodeConfig := range network.networkConfiguration.Nodes {
		if nodeConfig.IsBootstrapNode() {
			continue
		}
		_, checker, err := network.avalancheNetwork.CreateNodeNoCheck(network.networkConfiguration, nodeConfig)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating new node")
		}
		nonBootstrapNodeCheckers[id] = checker
	}
	logrus.Info("Non-bootstrap nodes launched")

	logrus.Info("Waiting for non-bootstrap nodes to become available...")
	for id, checker := range nonBootstrapNodeCheckers {
		if err := checker.WaitForStartup(timeBetweenNodeStartupPolls, maxNumNodeStartupPolls); err != nil {
			return stacktrace.Propagate(err, "An error occurred waiting for non-bootstrapper node '%v' to become available", id)
		}
	}
	logrus.Info("Non-bootstrap nodes available")

	topo := topology.New(network.avalancheNetwork)

	firstValidatorId := ""
	for _, node := range network.networkConfiguration.Nodes {
		topo.AddNode(node.ID, node.ID, constants.DefaultPassword)
		if firstValidatorId == "" {
			firstValidatorId = node.ID
		}
	}
	topo.AddGenesis(
		firstValidatorId,
		testconstants.GenesisUsername,
		testconstants.GenesisPassword,
	)

	firstNodeId := getBootstrapNodeId(initialBootstrapperIdIdx)
	firstNode := topo.Node(firstNodeId)
	firstNodeAvalancheGoClient := firstNode.GetClient()

	logrus.Info("Creating C-Chain address...")
	firstNodeCChainApi := firstNodeAvalancheGoClient.CChainAPI()
	// Client ...
	type Client struct {
		requester rpc.EndpointRequester
	}
	xChainPrivateKey, err := firstNodeAvalancheGoClient.XChainAPI().ExportKey(firstNode.UserPass, firstNode.XAddress)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred exporting the node's private key from the X-Chain")
	}
	logrus.Debugf("X-Chain private key: %v", xChainPrivateKey)
	cChainAddrStr, err := firstNodeAvalancheGoClient.CChainAPI().ImportKey(firstNode.UserPass, xChainPrivateKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred importing the node's private key to the C-Chain")
	}
	logrus.Infof("C-Chain address '%v' created", cChainAddrStr)

	logrus.Info("Transferring balance to C-Chain address...")
	cChainAddrBytes := common.HexToAddress(cChainAddrStr)
	genesis := topo.Genesis()
	genesis.MoveBalanceToCChain(cChainAddrBytes, testconstants.TxFee)
	logrus.Info("Balance transferred to C-Chain address")

	logrus.Info("Creating keyed transactor...")
	_, privKeyHexWithLead0x, err := firstNodeCChainApi.ExportKey(firstNode.UserPass, cChainAddrStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred exporting the private key for the C-Chain address '%v'", cChainAddrStr)
	}
	privKeyHex := strings.Replace(privKeyHexWithLead0x, hexStrIndicatorLeader, "", 1)
	logrus.Infof("C-Chain private key hex: %v", privKeyHex)
	privKeyEcdsa, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred converting the C-Chain private key hex '%v' to an ECDSA key", privKeyHex)
	}
	transactor := bind.NewKeyedTransactor(privKeyEcdsa)
	logrus.Info("Keyed transactor created")


	logrus.Info("Creating Geth client...")
	uri := fmt.Sprintf("ws://%s:%d/ext/bc/C/ws", firstNode.GetIPAddress(), nodeHttpPort)
	gethClient, err := ethclient.Dial(uri)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an ethclient for URI '%v'", uri)
	}
	logrus.Info("Geth client created")

	network.transactor = transactor
	network.gethClient = gethClient

	return nil
}

// Returns a Geth client for interacting with the Avalanche network's C-Chain, as well as a transactor required for making
//  requests using the Geth client
func (network SmartContractAvalancheNetwork) GetFundedCChainClientAndTransactor() (*ethclient.Client, *bind.TransactOpts) {
	return network.gethClient, network.transactor
}

func getBootstrapNodeId(idx int) string {
	return fmt.Sprintf("bootstrapNode-%d", idx)
}

