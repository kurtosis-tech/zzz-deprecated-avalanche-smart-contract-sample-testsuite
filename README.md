Example Avalanche Smart Contract Testsuite
==========================================
This testsuite demonstrates using Kurtosis to test functionality of a smart contract deployed on the Avalanche network C-Chain. 

### Running the testsuite
1. Install `git` on your machine if not done already
1. Install `docker` on your machine if not done already
1. Run `scripts/build-and-run.sh all`

### Uploading a new smart contract and regenerating the Go bindings
1. Install `go` on your machine if not done already
1. Install `solc` v0.7 on your machine (NOTE: **not** v0.8, which is the latest! This requirement is because the AvalancheGo client depends on an old version of `go-ethereum`):
    1. Install the command:
        * On Mac, this can be done via `brew tap ethereum/ethereum && brew install solidity@7`
        * On Linux, (untested) guidance is here: https://docs.soliditylang.org/en/v0.8.0/installing-solidity.html#linux-packages
    1. Verify that your version is v0.7: `solc --version`
1. Make the `abigen` binary using `go-ethereum` v1.9.21 on your local machine (NOTE: v1.9.21 is important for the same AvalancheGo reason):
    1. Clone the `go-ethereum` repo at v1.9.21: `git clone --branch v1.9.21 git@github.com:ethereum/go-ethereum.git`
    1. Enter the repo: `cd go-ethereum`
    1. Install all the binaries: `go install ./...`
    1. Verify that the installed `abigen` binary is on the correct version: `$GOPATH/bin/abigen --version`
1. Inside the testsuite repo, copy your contract to `smart_contracts/contract.sol`
1. Run the script to regenerate the Go bindings: `scripts/regenerate-contract-bindings.sh $GOTPATH/bin/abigen`
1. Fix the compile errors that occur in `smart_contract_test_.go` due to your new contract bindings
1. Verify the testsuite still works: `scripts/build-and-run.sh all`
