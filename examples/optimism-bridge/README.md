# Optimism L1 Standard Bridge implementation change

The OP Mainnet registry identifies `0x99C9fc46f92E8a1c0deC1b1747d010903E884bE1` as the L1 Standard Bridge proxy. These snapshots record its storage on Ethereum mainnet. [OP Mainnet registry](https://github.com/ethereum-optimism/superchain-registry/blob/main/superchain/configs/mainnet/op.toml)

| ERC-1967 slot | Block 22,182,530 | Block 22,182,531 |
| --- | --- | --- |
| Implementation | `0x64b5a5ed26dcb17370ff4d33a8d503f0fbd06cff` | `0x78972e88ab8bbb517a36caea23b931bab58ad3c6` |
| Administrator | `0x543ba4aadbab8f9025686bd03993043599c6fb04` | `0x543ba4aadbab8f9025686bd03993043599c6fb04` |

The evidence shows a change in the implementation slot between consecutive end-of-block states. It leaves transaction attribution and effective control of the administrator address open.

## Reproduce offline

From the repository root, build the binary as described in the [project README](../../README.md). Cross-check the block hashes against an independently trusted Ethereum source before relying on the result. The archived proofs were collected from MEV Blocker. [Ethereum block 22,182,530](https://etherscan.io/block/22182530) · [Ethereum block 22,182,531](https://etherscan.io/block/22182531)

```sh
./govarchive diff \
  -address 0x99c9fc46f92e8a1c0dec1b1747d010903e884be1 \
  -block 0x1df0e378e2618cd0a40cbfc1d6faeb2dff3de609a1c198e6a2534aae84b04794 \
  -next-block 0x379caf7ef6649d359d71f115b4e6ff3cd549bd05ba48b1e644b6924ed93aa13d \
  examples/optimism-bridge/before.json examples/optimism-bridge/after.json
```

Expect one changed slot named `implementation`. Change a claimed `storageProof` value in a copy of either bundle and rerun the command with that copy to see verification fail.
