# Optimism Upgrade 8 raised the Security Council signing threshold

On June 10, 2024 at 18:48:23 UTC, the Optimism Security Council Safe at `0xc2819dc788505aac350142a7a707bf9d03e3bd03` emitted `ChangedThreshold(10)`. The surrounding storage proofs show that the owner-signature path went from requiring **4 of 13** owners to **10 of 13**. [Governance proposal](https://gov.optimism.io/t/final-protocol-upgrade-8-guardian-security-council-threshold-and-l2-proxyadmin-ownership-changes-for-stage-1-decentralization/8157) · [Execution receipt](https://etherscan.io/tx/0x50db9c1b9f5a8616f982022429a1d75c35986bcfb67bd1e3358d97ddbe751144#eventlog)

| Safe storage | [Block 20,063,318](https://etherscan.io/block/20063318) | [Block 20,063,319](https://etherscan.io/block/20063319) |
| --- | --- | --- |
| Singleton, slot 0 | `0xd9db270c1b5e3bd161e8c8503c55ceabee709552` | Same address |
| Owner count, slot 3 | 13 | 13 |
| Threshold, slot 4 | 4 | 10 |

The post-execution proofs connect the [L1 bridge](control/bridge.json) to its [ProxyAdmin](control/proxy-admin.json), then to the [2-of-2 upgrade Safe](control/upgrade-safe.json). The upgrade Safe's owner list includes the Security Council Safe and the Foundation Safe. Along the owner-signature path, obtaining the Council's approval now required six more owner confirmations. This case verifies the owner count, threshold and post-execution route. The proof set does not cover individual Council owner changes or alternate module paths. [Safe v1.3.0 storage layout](https://github.com/safe-global/safe-smart-account/blob/v1.3.0/contracts/examples/libraries/GnosisSafeStorage.sol)

## Reproduce offline

Build the CLI from the repository root. Cross-check both block hashes independently. The storage proofs were collected from MEV Blocker.

```sh
COUNCIL=0xc2819dc788505aac350142a7a707bf9d03e3bd03
BEFORE=0xaf620161265847641248b219dc681cab3cfd19e7777414c145da3b2e55ecd01d
AFTER=0x5989cc0cff78aa28e51a5bd9bd80dcae01f39822a93e0004446916245dcacfbb
CASE=examples/optimism-council-threshold

./govarchive verify -address "$COUNCIL" -block "$BEFORE" "$CASE/before.json"
./govarchive verify -address "$COUNCIL" -block "$AFTER" "$CASE/after.json"
```

Read the storage word at key `0x0000000000000000000000000000000000000000000000000000000000000004`. It ends in `04` before and `0a` after. To independently examine the event itself, use the linked receipt.
