package smart_contract_test

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/smart_contracts/bindings"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/networks_impl"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math/big"
	"time"
)

const (
	maxNumCheckTransactionMinedRetries = 10
	timeBetweenCheckTransactionMinedRetries = 1 * time.Second
)

type SmartContractTest struct {
	avalancheImage string
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

func (test SmartContractTest) GetSetupTimeout() time.Duration {
	return 60 * time.Second
}

func (test SmartContractTest) GetExecutionTimeout() time.Duration {
	return 60 * time.Second
}

func (test *SmartContractTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	network := networks_impl.NewSmartContractAvalancheNetwork(test.avalancheImage)
	if err := network.ExecuteSetupPhaseInitialization(networkCtx); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing the setup phase initialization")
	}
	return network, nil
}

func (test SmartContractTest) Run(uncastedNetwork networks.Network, testCtx testsuite.TestContext) {
	// Necessary because Go doesn't have generics
	network, ok := uncastedNetwork.(*networks_impl.SmartContractAvalancheNetwork)
	if !ok {
		testCtx.Fatal(stacktrace.NewError("Couldn't cast the generic network to the appropriate type"))
	}
	transactor, gethClient, err := network.ExecuteRunPhaseInitialization(testCtx)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred executing the run phase initialization"))
	}

	logrus.Info("Deploying contract...")
	_, contractDeploymentTxn, contract, err := bindings.DeployStorage(transactor, gethClient)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred deploying the contract on the C-Chain"))
	}
	if err := waitUntilTransactionMined(gethClient, contractDeploymentTxn.Hash()); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred waiting for the contract deployment transaction to be mined"))
	}
	logrus.Info("Contract deployed")

	valueToStore := big.NewInt(20)
	logrus.Infof("Storing value '%v'...", valueToStore)
	storeValueTxn, err := contract.Store(transactor, valueToStore)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred storing value '%v' in the contract", valueToStore))
	}
	if err := waitUntilTransactionMined(gethClient, storeValueTxn.Hash()); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred waiting for the value-storing transaction to be mined"))
	}
	logrus.Info("Value stored")

	// NOTE: It's not clear why we need to sleep here - the transaction being mined should be sufficient
	time.Sleep(5 * time.Second)

	logrus.Info("Retrieving value from contract...")
	retrievedValue, err := contract.Retrieve(&bind.CallOpts{})
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred retrieving the value stored in the contract"))
	}
	logrus.Infof("Retrieved value: %v", retrievedValue)

	testCtx.AssertTrue(valueToStore.Cmp(retrievedValue) == 0, stacktrace.NewError("Retrieved value '%v' != stored value '%v'", retrievedValue, valueToStore))
}


// If we try to use a contract immediately after submission without waiting for it to be mined, we'll get a "no contract code at address" error:
// https://github.com/ethereum/go-ethereum/issues/15930#issuecomment-532144875
func waitUntilTransactionMined(validatorClient *ethclient.Client, transactionHash common.Hash) error {
	for i := 0; i < maxNumCheckTransactionMinedRetries; i++ {
		receipt, err := validatorClient.TransactionReceipt(context.Background(), transactionHash)
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
