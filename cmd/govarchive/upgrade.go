package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"

	"github.com/legendinn/ethereum-governance-archive/internal/evidence"
)

func verifyUpgrade(args []string, out, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("verify-upgrade", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	blockFlag := flags.String("block", "", "independently trusted block hash")
	txFlag := flags.String("tx", "", "expected transaction hash")
	proxyFlag := flags.String("address", "", "expected proxy address")
	implementationFlag := flags.String("implementation", "", "expected new implementation address")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("verify-upgrade requires one proof file")
	}
	var block, tx common.Hash
	if err := block.UnmarshalText([]byte(*blockFlag)); err != nil {
		return fmt.Errorf("invalid -block: %w", err)
	}
	if err := tx.UnmarshalText([]byte(*txFlag)); err != nil {
		return fmt.Errorf("invalid -tx: %w", err)
	}
	if !common.IsHexAddress(*proxyFlag) || !common.IsHexAddress(*implementationFlag) {
		return fmt.Errorf("-address and -implementation must contain 20-byte Ethereum addresses")
	}
	var proof evidence.UpgradeProof
	if err := readBundle(flags.Arg(0), &proof); err != nil {
		return err
	}
	if err := proof.Verify(block, tx, common.HexToAddress(*proxyFlag), common.HexToAddress(*implementationFlag)); err != nil {
		return err
	}
	_, err := fmt.Fprintf(out, "Verified Upgraded event in transaction %s at block index %d\n", tx, *proof.Index)
	return err
}
