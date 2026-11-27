package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/legendinn/ethereum-governance-archive/internal/evidence"
)

var proxySlots = []struct {
	name string
	key  common.Hash
}{
	{"implementation", common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")},
	{"admin", common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")},
}

func runDiff(args []string, out, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("diff", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	var target snapshotFlags
	target.register(flags)
	nextBlock := flags.String("next-block", "", "second independently trusted block hash")
	if err := flags.Parse(args); err != nil {
		return err
	}
	address, before, err := target.parse()
	if err != nil {
		return err
	}
	if flags.NArg() != 2 {
		return fmt.Errorf("diff requires two bundle files")
	}
	var after common.Hash
	if err := after.UnmarshalText([]byte(*nextBlock)); err != nil {
		return fmt.Errorf("-next-block must contain a 32-byte block hash: %w", err)
	}
	return diff(out, address, before, after, flags.Arg(0), flags.Arg(1))
}

func diff(out io.Writer, account common.Address, before, after common.Hash, firstPath, secondPath string) error {
	first, err := readSnapshot(firstPath, before, account)
	if err != nil {
		return err
	}
	second, err := readSnapshot(secondPath, after, account)
	if err != nil {
		return err
	}
	oldValues, err := proxyValues(first)
	if err != nil {
		return err
	}
	newValues, err := proxyValues(second)
	if err != nil {
		return err
	}
	type change struct {
		Slot   string         `json:"slot"`
		Before common.Address `json:"before"`
		After  common.Address `json:"after"`
	}
	changes := make([]change, 0, len(proxySlots))
	for i, slot := range proxySlots {
		if oldValues[i] != newValues[i] {
			changes = append(changes, change{slot.name, oldValues[i], newValues[i]})
		}
	}
	return writeJSON(out, struct {
		Account common.Address `json:"account"`
		Before  common.Hash    `json:"beforeBlock"`
		After   common.Hash    `json:"afterBlock"`
		Changes []change       `json:"changes"`
	}{account, before, after, changes})
}

func proxyValues(snapshot evidence.Snapshot) ([]common.Address, error) {
	values := make([]common.Address, 0, len(proxySlots))
	for _, slot := range proxySlots {
		word, found := snapshot.Slots[slot.key]
		if !found {
			return nil, fmt.Errorf("missing %s slot at %s", slot.name, snapshot.BlockHash)
		}
		addressBytes := word[common.HashLength-common.AddressLength:]
		if common.BytesToHash(addressBytes) != word {
			return nil, fmt.Errorf("%s slot does not encode an address", slot.name)
		}
		values = append(values, common.BytesToAddress(addressBytes))
	}
	return values, nil
}
