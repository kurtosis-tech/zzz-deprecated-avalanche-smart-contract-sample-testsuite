set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# Constants
SMART_CONTRACTS_DIRNAME="smart_contracts"
BINDINGS_DIRNAME="bindings"
CONTRACT_RELATIVE_FILEPATH="${SMART_CONTRACTS_DIRNAME}/contract.sol"   # Relative to repo root
BINDINGS_CODE_RELATIVE_FILEPATH="${SMART_CONTRACTS_DIRNAME}/${BINDINGS_DIRNAME}/bindings.go"  # Relative to repo root
REQUIRED_SOLIDITY_VERSION="0.7" # This is fixed to 0.7 because the Avalanche Kurtosis bindings use ethereum-go 0.7 and newer versions break

# Main code
if [ "${#}" -ne 1 ]; then
    echo "Usage: $(basename "${0}") /path/to/v${REQUIRED_SOLIDITY_VERSION}/abigen/binary"
    exit 1
fi
abigen_binary_filepath="${1}"

if ! command -v solc; then
    echo "Error: Solidity v${REQUIRED_SOLIDITY_VERSION} must be installed" >&2
    exit 1
fi
solidity_version="$(solc --version | tail -1 | awk '{print $2}')"
case "${solidity_version}" in
    ${REQUIRED_SOLIDITY_VERSION}*)
        # Version matches
        ;;
    *)
        echo "Error: Installed version of Solidity is '${solidity_version}' but must be ${REQUIRED_SOLIDITY_VERSION}"
        exit 1
        ;;
esac

contract_filepath="${root_dirpath}/${CONTRACT_RELATIVE_FILEPATH}"
bindings_filepath="${root_dirpath}/${BINDINGS_CODE_RELATIVE_FILEPATH}"
if ! "${abigen_binary_filepath}" --sol "${contract_filepath}" --pkg "${BINDINGS_DIRNAME}" --out "${bindings_filepath}"; then
    echo "Error: Could not generate bindings for Solidity contract at '${contract_filepath}'" >&2
    exit 1
fi
echo "Successfully generated bindings for Solidity contract to file '${bindings_filepath}'"
