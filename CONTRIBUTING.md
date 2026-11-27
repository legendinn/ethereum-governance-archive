# Contributing

## Add a historical case

1. Choose a protocol change with an independently identifiable Ethereum block, account, and governance source. Record the block hashes alongside the block numbers.
2. Capture the relevant storage before and after the change. For a control claim, include the storage proof and contract source that explain each link in the chain of authorization.
3. Add a directory under `examples/` with the proof files and a README. State what changed, how the action was authorized, what users could observe, and how to reproduce the evidence checks. Link the original proposal and execution record where available.
4. Add the case to the root README. When a case introduces a new proof or verification path, add a regression test with independently pinned inputs.

The [Compound case](examples/compound-usdc/README.md) shows how to present a timelocked upgrade. The [Optimism Council case](examples/optimism-council-threshold/README.md) shows how to document a change in authority.

## Validate a change

```sh
gofmt -l cmd internal
go test ./...
go vet ./...
```

For evidence changes, run the verification commands from the case README and record the result in the pull request. Keep RPC credentials and local configuration out of commits. For factual corrections, include the claim, the block reference, and the source supporting the correction.
