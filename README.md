# Ethereum Governance Archive

A historical archive for DeFi risk analysts, smart-contract auditors, and governance researchers. It records **what changed, who could authorize it, and how much time users had to react**. Each case includes Ethereum evidence you can verify offline.

## Try the Compound demo

From the repository root, with Go 1.25+ and Python 3 installed:

```sh
go build -o govarchive ./cmd/govarchive
./scripts/demo-compound.sh
```

The CLI verifies that the Compound III USDC proxy emitted `Upgraded` in the named Ethereum transaction. The demo then alters the implementation address in a temporary copy of the receipt proof and confirms that verification fails. You need network access to download Go dependencies on the first build. The demo itself runs offline. [Full case and standalone verification command](examples/compound-usdc/README.md#verify-the-execution-event-offline)

## Browse the archive

| Case | What changed | Governance evidence |
| --- | --- | --- |
| [Compound III · Ethereum USDC](examples/compound-usdc/README.md) | New Comet implementation | Governor Bravo and a 48-hour timelock. Execution followed queueing by 48 hours and 2 minutes. |
| [Optimism · Upgrade 13](examples/optimism-bridge/README.md) | New L1 bridge implementation | A 2-of-2 Safe owned by the Foundation and the Security Council. The observed execution path had no on-chain timelock. |
| [Optimism · Upgrade 8](examples/optimism-council-threshold/README.md) | Security Council threshold rose from 4 of 13 to 10 of 13 | Consecutive block proofs record the threshold change. Post-upgrade proofs connect the Council Safe to the bridge’s authorization path. |

## Use the CLI

| Command | Purpose | Network |
| --- | --- | --- |
| `capture` | Fetch and verify Ethereum storage proofs for a block | Historical Ethereum execution RPC |
| `verify` | Authenticate storage values against a trusted block hash | Offline |
| `diff` | Compare implementation and admin slots at two trusted blocks | Offline |
| `verify-upgrade` | Authenticate a transaction and its proxy’s `Upgraded` event | Offline |

Verification starts with a block hash you trust from an independent source. Proofs authenticate storage and, where supplied, receipt contents against that block. Case studies cite traces, receipts, and contract source for claims about execution and governance. [CLI guide](docs/cli.md) · [Evidence formats and trust model](docs/evidence.md)

## Develop and contribute

```sh
go test ./...
go vet ./...
```

### Repository layout

| Directory | Responsibility |
| --- | --- |
| [`cmd/govarchive`](cmd/govarchive) | CLI flags, bundle files, and ERC-1967 reports |
| [`internal/ethereumrpc`](internal/ethereumrpc) | Collects storage proofs from Ethereum RPC and checks that they match the request |
| [`internal/evidence`](internal/evidence) | Verifies storage, transaction, and receipt proofs without a network connection |
| [`examples`](examples) | Historical case studies and proof fixtures used by the tests |

Only `capture` calls the RPC layer. Both captured and archived files use the same proof verifier. GitHub Actions checks formatting, module files, tests, `go vet`, and the offline demo on Go 1.25.

[How to contribute a case](CONTRIBUTING.md) · [MIT license](LICENSE)
