package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/legendinn/ethereum-governance-archive/internal/evidence"
)

type snapshotFlags struct {
	address string
	block   string
}

func (s *snapshotFlags) register(flags *flag.FlagSet) {
	flags.StringVar(&s.address, "address", "", "expected Ethereum account address")
	flags.StringVar(&s.block, "block", "", "independently trusted block hash")
}

func (s snapshotFlags) parse() (common.Address, common.Hash, error) {
	if !common.IsHexAddress(s.address) {
		return common.Address{}, common.Hash{}, fmt.Errorf("-address must contain a 20-byte Ethereum address")
	}
	var block common.Hash
	if err := block.UnmarshalText([]byte(s.block)); err != nil {
		return common.Address{}, common.Hash{}, fmt.Errorf("-block must contain a 32-byte block hash: %w", err)
	}
	return common.HexToAddress(s.address), block, nil
}

func runVerify(args []string, out, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	var target snapshotFlags
	target.register(flags)
	if err := flags.Parse(args); err != nil {
		return err
	}
	address, block, err := target.parse()
	if err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("verify requires one bundle file")
	}
	snapshot, err := readSnapshot(flags.Arg(0), block, address)
	if err != nil {
		return err
	}
	return writeJSON(out, snapshot)
}

func readBundle(path string, value any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	const maxBundle = 4 << 20
	data, err := io.ReadAll(io.LimitReader(file, maxBundle+1))
	if err != nil {
		return err
	}
	if len(data) > maxBundle {
		return fmt.Errorf("%s exceeds the 4 MiB bundle limit", path)
	}
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func readSnapshot(path string, anchor common.Hash, address common.Address) (evidence.Snapshot, error) {
	var bundle evidence.Bundle
	if err := readBundle(path, &bundle); err != nil {
		return evidence.Snapshot{}, err
	}
	snapshot, err := bundle.Verify(anchor, address)
	if err != nil {
		return evidence.Snapshot{}, fmt.Errorf("%s: %w", path, err)
	}
	return snapshot, nil
}

func writeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
