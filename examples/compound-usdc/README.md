# Compound III Ethereum USDC implementation change

Compound proposal 421 changed the implementation of the Ethereum USDC Comet proxy at `0xc3d688b66703497daa19211eedff47f25384cdc3`. The proxy emitted `Upgraded` in the execution receipt with the new implementation address. [Compound deployment and governance docs](https://docs.compound.finance/governance/) · [Execution receipt](https://etherscan.io/tx/0xf9df4d7016a592d1135dee34cacae43030c27c3eb69a12281385faa6bd487f56#eventlog)

| ERC-1967 slot | Block 22,167,541 | Block 22,167,542 |
| --- | --- | --- |
| Implementation | `0xaec1954467b6d823a9042e9e9d6e4f40111069a9` | `0x2c045e9a1afb125898d6e736d1ff09fa28788e9d` |
| Administrator | `0x1ec63b5883c3481134fd50d5daebc83ecd2e8779` | `0x1ec63b5883c3481134fd50d5daebc83ecd2e8779` |

## Execution and notice

| Event | Ethereum block | UTC |
| --- | --- | --- |
| [Proposal queued](https://etherscan.io/tx/0xd3a72b49052d517015261378e2ff64c504a27da4df1507524536e9abaf8719ee#eventlog) | 22,153,198 | March 29, 2025, 14:36:35 |
| Earliest execution per `ProposalQueued` | | March 31, 2025, 14:36:35 |
| [Proposal executed](https://etherscan.io/tx/0xf9df4d7016a592d1135dee34cacae43030c27c3eb69a12281385faa6bd487f56#eventlog) | 22,167,542 | March 31, 2025, 14:38:35 |

The observed notice after queueing was 48 hours and 2 minutes. The RPC call trace shows the Governor calling the Timelock, the Timelock calling `CometProxyAdmin.deployAndUpgradeTo`, and the ProxyAdmin calling the Comet proxy's `upgradeTo` with the new implementation. Use `debug_traceTransaction` with `{"tracer":"callTracer"}` on a trace-capable RPC to reproduce the call path. The trace was obtained from dRPC. [Compound governance architecture](https://docs.compound.finance/governance/) · [CometProxyAdmin source](https://github.com/compound-finance/comet/blob/main/contracts/CometProxyAdmin.sol)

| Contract | Ethereum address | State immediately before execution |
| --- | --- | --- |
| [CometProxyAdmin proof](control/proxy-admin.json) | `0x1ec63b5883c3481134fd50d5daebc83ecd2e8779` | Slot 0 identifies the Timelock as owner |
| [Timelock proof](control/timelock.json) | `0x6d903f6003cca6255d85cca4d3b5e5146dc33925` | Slot 0 identifies Governor Bravo at `0x309a862bbc1a00e45506cb8a802d1ff10004c8c0` as admin. Slot 2 is `0x2a300` or 172,800 seconds (48 hours). |

The offline proofs establish the recorded contract state at the anchored blocks. The linked queue and execution receipts and the RPC trace establish the execution timeline and call path.

## Reproduce offline

The proofs were collected from MEV Blocker. Cross-check the hashes with an independently trusted Ethereum source before relying on the results. [Ethereum block 22,167,541](https://etherscan.io/block/22167541) · [Ethereum block 22,167,542](https://etherscan.io/block/22167542)

```sh
./govarchive diff \
  -address 0xc3d688b66703497daa19211eedff47f25384cdc3 \
  -block 0x5caa47372cefc6b87a7b0e2b18d30b9ab9581ff1f9a427bfc3800f63cb1afae2 \
  -next-block 0xdd782654955f5b0ffb41ab185b43a92e7220d0c07901fb29ee676577956c5a72 \
  examples/compound-usdc/before.json examples/compound-usdc/after.json

./govarchive verify \
  -address 0x6d903f6003cca6255d85cca4d3b5e5146dc33925 \
  -block 0x5caa47372cefc6b87a7b0e2b18d30b9ab9581ff1f9a427bfc3800f63cb1afae2 \
  examples/compound-usdc/control/timelock.json
```
