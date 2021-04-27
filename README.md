Example Avalanche Smart Contract Testsuite
==========================================
This testsuite demonstrates using [Kurtosis](https://www.kurtosistech.com/) to test functionality of a smart contract deployed on the Avalanche network C-Chain. 

Kurtosis docs are available [here](https://docs.kurtosistech.com/), and the team is available on [Discord](https://discord.gg/6Jjp9c89z9).

1 - Create your own repo from this repository
---------------------------------------------
1. Clone this repository to your local machine
1. Delete the `.git` directory: `rm -rf .git`
1. Reinitialize Git: `git init`
1. Add all the files in this repo: `git add .`
1. Commit: `git commit -m "Initial commit"`
1. Create a new repo on Github
1. Add your new repo as a remote: `git remote add origin git@github.com:YOUR-ORG-HERE/YOUR-REPO-HERE.git`
1. Push up to your repo: `git push origin`

2 - Run the testsuite
---------------------
1. Install `git` on your machine if not done already
1. Install `docker` on your machine if not done already
1. Create a Kurtosis account [here](https://www.kurtosistech.com/sign-up)
1. Run `scripts/build-and-run.sh all`, and when prompted link your device to your Kurtosis account

3 - Upload your smart contracts and regenerate the Go bindings
--------------------------------------------------------------
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

4 - Customize the testsuite
---------------------------
1. Install `go` on your machine
1. Install a Go IDE of your choice (we recommend [GoLand by JetBrains](https://www.jetbrains.com/go/))
1. Open the repo directory
1. Replace the section marked `TODO REPLACE WITH YOUR TEST CODE` in `testsuite/testsuite_impl/smart_contract_test_.go` using the bindings generated for your contracts
1. Verify the testsuite still works: `scripts/build-and-run.sh all`
1. Add more tests as you please
    * The [Testsuite Customization](https://docs.kurtosistech.com/kurtosis-core/testsuite-customization) docs page provides a step-by-step walkthrough to customizing a testsuite
    * The [Lib Documentation](https://docs.kurtosistech.com/kurtosis-libs/lib-documentation) docs page provides comprehensive documentation of all the Kurtosis components you'll encounter
