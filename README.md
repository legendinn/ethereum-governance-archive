# Ethereum Governance Archive

Download and verify historical evidence about Ethereum proxy changes. The first case tracks an implementation change at the Ethereum L1 Standard Bridge proxy for OP Mainnet across blocks 22,182,530 and 22,182,531. The administrator slot has the same value at both blocks.

[Read the case study and reproduce the diff](examples/optimism-bridge/README.md).

## Build

Requires Go 1.25 or later. The first build downloads dependencies. Verification and comparison run offline using the built binary.

```sh
go test ./...
go build -o govarchive ./cmd/govarchive
```

## Capture another snapshot

Set `ETH_RPC_URL` to an Ethereum execution RPC with historical proof support. Find the expected proxy address and canonical block hash independently. The endpoint must accept an EIP-1898 block hash selector for `eth_getProof`.

```sh
./govarchive capture -address "$PROXY" -block "$BLOCK_HASH" > snapshot.json
./govarchive verify -address "$PROXY" -block "$BLOCK_HASH" snapshot.json
```

To compare two verified snapshots of the same proxy, pass the earlier and later block hashes to `diff` as `-block` and `-next-block`, followed by the two bundle paths.

## Evidence and trust

A version 1 bundle contains an execution block header, an account proof, and storage proofs with their claimed values. The verifier hashes the header and compares it with the caller's trusted block hash. It verifies the account against the header's state root, then verifies each storage claim against the storage root proven for that account. Bundle contents alone cannot establish whether a block is canonical.

Capture and diff cover the ERC-1967 implementation and administrator slots. The report describes end-of-block storage. Control of an administrator contract, timelock delays, upgrade authorization, and transaction attribution each require additional evidence. The archive currently contains one manually selected case.

- `cmd/govarchive` contains the CLI, RPC capture, and ERC-1967 comparison.
- `internal/evidence` contains the versioned bundle format and offline verification.
- `examples` contains case studies, trusted block references to cross-check, and downloadable proof bundles.

[ERC-1967](https://eips.ethereum.org/EIPS/eip-1967) · [Execution API for account and storage proofs](https://ethereum.github.io/execution-apis/api/methods/eth_getProof/)
