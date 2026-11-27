# Evidence formats and trust model

Every verification starts with a block hash supplied by the caller. For real-world use, obtain it from a source you trust to identify the canonical Ethereum chain, such as your own node. The verifier recomputes the hash of the execution header before using the committed roots.

## Storage bundle, version 1

| Field | Meaning |
| --- | --- |
| `version` | Integer `1` |
| `header` | Ethereum execution block header in RPC JSON form |
| `account.address` | Address to be matched against the caller’s expected account |
| `account.accountProof` | Hex-encoded account-trie nodes |
| `account.storageProof` | Entries containing `key`, `value`, and `proof` for each slot |

The verifier checks the account proof against the header’s state root. It decodes the proven account’s storage root and checks each requested word against that root. An absent storage entry can prove a zero value. Verification describes the stored words at the end of that block. A case study must cite the layout used to interpret those words.

[Account and storage proofs](https://ethereum.github.io/execution-apis/api/methods/eth_getProof/) · [ERC-1967 slots](https://eips.ethereum.org/EIPS/eip-1967)

## Upgrade bundle, version 1

| Field | Meaning |
| --- | --- |
| `version` | Integer `1` |
| `header` | Ethereum execution block header in RPC JSON form |
| `index` | Zero-based transaction position in the block |
| `transactionProof` | Hex-encoded transaction-trie nodes |
| `receiptProof` | Hex-encoded receipt-trie nodes |

The verifier checks both tries at the same RLP-encoded index, derives the expected transaction hash from the proven transaction bytes, and checks that the matching receipt succeeded. It looks for `Upgraded(address)` emitted by the expected proxy with the expected indexed implementation address.

## Interpreting the result

A proof binds data to the selected block. Establishing chain canonicality, attributing an internal call path, and interpreting Safe or timelock storage require their respective evidence. A proven event establishes that an emitter logged a particular value. To establish which value ended up in storage, use a separate storage proof. When writing a case, label any claim based on an RPC trace or project documentation and link the source.
