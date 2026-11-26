# Compound case study review

Use this 15-minute session to see whether the case study helps a DeFi risk analyst support an exposure review. Give the reviewer the [Compound case study](../examples/compound-usdc/README.md), then let them read it and answer before sharing the facilitator notes. The repository is private. Confirm that the reviewer can open the case study before scheduling the session.

## Invitation

I'm building an archive that helps DeFi risk teams check who can upgrade a protocol and how much notice they get. Could you spend 15 minutes reviewing one historical Compound III upgrade? I'd like to see which evidence you would cite in an exposure review and what you would still need to verify. There's no setup required for the reading portion.

## Reviewer tasks

1. Say which contracts authorized and executed the upgrade. Show where you found that information.
2. Work out how much time elapsed from queueing to execution. Explain how that differs from the minimum enforced delay.
3. Select the evidence you would cite to support a review of the market. Identify one claim you can check offline and one you would verify separately.
4. Describe any missing evidence that would change your decision, and say which section you would return to for a future upgrade.

## Facilitator notes

- Look for the path from Governor Bravo through the Timelock and CometProxyAdmin to the proxy. Record which source the reviewer uses for the path.
- The minimum delay was 48 hours. The observed interval was 48 hours and 2 minutes.
- The offline checks authenticate storage at specific blocks and the proxy's `Upgraded` event in the named transaction. The detailed call path comes from the execution trace. Queueing time comes from the chain records linked in the case study.
- Record the time and source for each answer, any unsupported inference, and the first question that requires leaving the case study. Ask what they would actually copy into their review. Publish feedback only with the reviewer's permission.
