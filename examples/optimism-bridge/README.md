# Optimism L1 Standard Bridge implementation change

The OP Mainnet registry identifies `0x99C9fc46f92E8a1c0deC1b1747d010903E884bE1` as the L1 Standard Bridge proxy. These snapshots record its storage on Ethereum mainnet. [OP Mainnet registry](https://github.com/ethereum-optimism/superchain-registry/blob/main/superchain/configs/mainnet/op.toml)

| ERC-1967 slot | Block 22,182,530 | Block 22,182,531 |
| --- | --- | --- |
| Implementation | `0x64b5a5ed26dcb17370ff4d33a8d503f0fbd06cff` | `0x78972e88ab8bbb517a36caea23b931bab58ad3c6` |
| Administrator | `0x543ba4aadbab8f9025686bd03993043599c6fb04` | `0x543ba4aadbab8f9025686bd03993043599c6fb04` |

## Execution and authorization

Optimism [published Upgrade 13 on March 7, 2025](https://gov.optimism.io/t/upgrade-proposal-13-opcm-and-incident-response-improvements/9739). Its operations repository records execution on April 2, 2025 at 16:50:23 UTC in [transaction `0x5f3530e5…e65c29`](https://etherscan.io/tx/0x5f3530e593bbac37c61dc5b7755b6a40c06c20c1a3a1b13fca5b7d00cde65c29). The RPC call trace shows the upgrade Safe delegating through Multicall3 and OP Contracts Manager, then calling `ProxyAdmin.upgrade(bridge, implementation)`. The ProxyAdmin called `setStorage(bytes32,bytes32)` on this legacy ChugSplash proxy with the implementation slot and the address above. Optimism's published validation independently lists that exact storage change. [Execution record](https://github.com/ethereum-optimism/superchain-ops/blob/bcf38604f89011977a8d0ace03139f5853747360/src/tasks/eth/000-opcm-upgrade-v200/README.md) · [Expected state change](https://github.com/ethereum-optimism/superchain-ops/blob/bcf38604f89011977a8d0ace03139f5853747360/src/tasks/eth/000-opcm-upgrade-v200/VALIDATION.md#L710-L721) · [ProxyAdmin source](https://github.com/ethereum-optimism/optimism/blob/op-contracts/v2.0.0-rc.1/packages/contracts-bedrock/src/universal/ProxyAdmin.sol)

| Contract | Ethereum address | State immediately before execution |
| --- | --- | --- |
| [ProxyAdmin proof](control/proxy-admin.json) | `0x543ba4aadbab8f9025686bd03993043599c6fb04` | Slot 0 identifies the upgrade Safe as owner |
| [Upgrade Safe proof](control/upgrade-safe.json) | `0x5a0aae59d09fccbddb6c6cceb07b7279367c3d2a` | 2 of 2, with the Foundation Safe and Security Council Safe as owners |
| [Foundation Safe proof](control/foundation-safe.json) | `0x847b5c174615b1b7fdf770882256e2d3e95b9d92` | 5 of 7 |
| [Security Council Safe proof](control/council-safe.json) | `0xc2819dc788505aac350142a7a707bf9d03e3bd03` | 10 of 13 |

All four proof bundles are anchored to block 22,182,530. The Safe bundles include the singleton pointer in slot 0, owner count in slot 3, and threshold in slot 4. The parent Safe bundle includes the three owner mapping entries, computed as `keccak256(abi.encode(owner, uint256(2)))` for the sentinel and each nested Safe. Those entries link the sentinel to the Foundation Safe, the Foundation Safe to the Council Safe, and the Council Safe back to the sentinel. [Safe v1.3.0 storage layout](https://github.com/safe-global/safe-smart-account/blob/v1.3.0/contracts/examples/libraries/GnosisSafeStorage.sol) · [Nested Safe validation](https://github.com/ethereum-optimism/superchain-ops/blob/bcf38604f89011977a8d0ace03139f5853747360/docs/NESTED-VALIDATION.md)

The observed execution path contains no on-chain timelock. The March 7 publication gives a historical notice date for this upgrade. The recorded Safe configuration describes the owner-signature path and leaves enabled modules and other administrative routes outside this case.

The offline verifier authenticates the recorded storage words. To reproduce transaction attribution, query `debug_traceTransaction` with the transaction hash and `{"tracer":"callTracer"}` on a trace-capable Ethereum RPC. The trace was obtained from dRPC and cross-checked against Optimism's operations record.

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

To verify a control bundle, use the same trusted block hash as `-block` above and the account address from the control table:

```sh
./govarchive verify \
  -address 0x5a0aae59d09fccbddb6c6cceb07b7279367c3d2a \
  -block 0x1df0e378e2618cd0a40cbfc1d6faeb2dff3de609a1c198e6a2534aae84b04794 \
  examples/optimism-bridge/control/upgrade-safe.json
```
