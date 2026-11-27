# CLI guide

Build from the repository root with `go build -o govarchive ./cmd/govarchive`. Each subcommand supports `-help`. Place flags before filenames.

## Capture historical storage

Set `ETH_RPC_URL` to an Ethereum execution RPC that serves historical proofs and supports an [EIP-1898](https://eips.ethereum.org/EIPS/eip-1898) block hash selector for `eth_getProof`. Keep authenticated endpoints in environment variables. Obtain the account address and block hash independently and set `ACCOUNT` and `BLOCK_HASH`.

```sh
./govarchive capture -address "$ACCOUNT" -block "$BLOCK_HASH" > snapshot.json
./govarchive verify -address "$ACCOUNT" -block "$BLOCK_HASH" snapshot.json
```

The default capture requests the ERC-1967 implementation and admin slots. Use one or more `-slot` flags for contract-specific storage. Each key must be a full 32-byte hex value. For example, this requests slot 4:

```sh
./govarchive capture -address "$ACCOUNT" -block "$BLOCK_HASH" \
  -slot 0x0000000000000000000000000000000000000000000000000000000000000004 \
  > threshold.json
```

If an RPC cannot serve an older proof, retry with a provider that supports archival trie queries. Use the same block hash for the retry.

## Compare verified layouts

Set `BEFORE_BLOCK` and `AFTER_BLOCK` to independently trusted block hashes. Supply two snapshots for the same account in the corresponding order.

```sh
./govarchive diff -address "$ACCOUNT" -block "$BEFORE_BLOCK" \
  -next-block "$AFTER_BLOCK" before.json after.json
```

`diff` reports changed ERC-1967 slots in JSON. For a custom slot, run `verify` for each snapshot and compare the authenticated storage words.

## Verify an upgrade event

```sh
./govarchive verify-upgrade -block "$BLOCK_HASH" -tx "$TX_HASH" \
  -address "$PROXY" -implementation "$IMPLEMENTATION" upgrade-proof.json
```

The expected transaction hash, emitter address, and implementation address are supplied by the caller. A mismatch, invalid trie proof, or malformed input returns a non-zero exit code. Each input file is limited to 4 MiB.

[Worked Compound example](../examples/compound-usdc/README.md#verify-the-execution-event-offline) · [Evidence formats and trust model](evidence.md)
