set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# Constants
SMART_CONTRACTS_DIRNAME="smart_contracts"
BINDINGS_DIRNAME="bindings"
SOLIDITY_DIRNAME="solidity"
GO_FILE_EXT=".go"
SOLIDITY_FILE_EXT=".sol"
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

bindings_dirpath="${root_dirpath}/${SMART_CONTRACTS_DIRNAME}/${BINDINGS_DIRNAME}"
if ! find "${bindings_dirpath}" -type f -name "*${GO_FILE_EXT}" -delete; then
    echo "Error: Could not remove existing Go files in bindings directory '${bindings_dirpath}'" >&2
    exit 1
fi
for contract_filepath in $(find "${root_dirpath}/${SMART_CONTRACTS_DIRNAME}/${SOLIDITY_DIRNAME}" -type f -name "*${SOLIDITY_FILE_EXT}"); do
    contract_filename="$(basename "${contract_filepath}")"
    bindings_filename="${contract_filename%%${SOLIDITY_FILE_EXT}}${GO_FILE_EXT}"
    output_filepath="${bindings_dirpath}/${bindings_filename}"
    if ! "${abigen_binary_filepath}" --sol "${contract_filepath}" --pkg "${BINDINGS_DIRNAME}" --out "${output_filepath}"; then
        echo "Error: Could not generate bindings for Solidity contract at '${contract_filepath}'" >&2
        exit 1
    fi
    echo "Successfully generated bindings for Solidity contract at '${contract_filepath}' to file '${output_filepath}'"
done
