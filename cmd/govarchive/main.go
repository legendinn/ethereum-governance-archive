package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/legendinn/ethereum-governance-archive/internal/evidence"
)

var proxySlots = []struct {
	name string
	key  common.Hash
}{
	{"implementation", common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")},
	{"admin", common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")},
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out, diagnostics io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: govarchive <capture|verify|diff|verify-upgrade> [flags]")
	}
	if args[0] == "verify-upgrade" {
		return verifyUpgrade(args[1:], out, diagnostics)
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	addressFlag := flags.String("address", "", "expected Ethereum account address")
	blockFlag := flags.String("block", "", "independently trusted block hash")
	var rpcFlag, nextBlock *string
	var requestedSlots []common.Hash
	switch args[0] {
	case "capture":
		rpcFlag = flags.String("rpc", "", "Ethereum RPC URL (defaults to ETH_RPC_URL)")
		flags.Func("slot", "32-byte storage key to capture (repeatable, defaults to ERC-1967 slots)", func(input string) error {
			var key common.Hash
			if err := key.UnmarshalText([]byte(input)); err != nil {
				return err
			}
			requestedSlots = append(requestedSlots, key)
			return nil
		})
	case "diff":
		nextBlock = flags.String("next-block", "", "second independently trusted block hash")
	case "verify":
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if !common.IsHexAddress(*addressFlag) {
		return fmt.Errorf("-address must contain a 20-byte Ethereum address")
	}
	address := common.HexToAddress(*addressFlag)
	var anchor common.Hash
	if err := anchor.UnmarshalText([]byte(*blockFlag)); err != nil {
		return fmt.Errorf("-block must contain a 32-byte block hash: %w", err)
	}
	switch args[0] {
	case "capture":
		if flags.NArg() != 0 {
			return fmt.Errorf("capture accepts flags only")
		}
		endpoint := *rpcFlag
		if endpoint == "" {
			endpoint = os.Getenv("ETH_RPC_URL")
		}
		return capture(out, endpoint, anchor, address, requestedSlots)
	case "verify":
		if flags.NArg() != 1 {
			return fmt.Errorf("verify requires one bundle file")
		}
		snapshot, err := readSnapshot(flags.Arg(0), anchor, address)
		if err != nil {
			return err
		}
		return writeJSON(out, snapshot)
	case "diff":
		if flags.NArg() != 2 {
			return fmt.Errorf("diff requires two bundle files")
		}
		var second common.Hash
		if err := second.UnmarshalText([]byte(*nextBlock)); err != nil {
			return fmt.Errorf("-next-block must contain a 32-byte block hash: %w", err)
		}
		return diff(out, address, anchor, second, flags.Arg(0), flags.Arg(1))
	}
	return nil
}

func capture(out io.Writer, endpoint string, anchor common.Hash, address common.Address, keys []common.Hash) error {
	if endpoint == "" {
		return fmt.Errorf("set ETH_RPC_URL or provide -rpc")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	client, err := rpc.DialContext(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("connect to RPC: %w", err)
	}
	defer client.Close()
	var header *types.Header
	if err := client.CallContext(ctx, &header, "eth_getBlockByHash", anchor, false); err != nil {
		return fmt.Errorf("get block header: %w", err)
	}
	if header == nil || header.Hash() != anchor {
		return fmt.Errorf("RPC did not return the requested block")
	}
	var account evidence.AccountProof
	defaultSlots := len(keys) == 0
	if defaultSlots {
		keys = []common.Hash{proxySlots[0].key, proxySlots[1].key}
	}
	expected := make(map[common.Hash]struct{}, len(keys))
	for _, key := range keys {
		if _, duplicate := expected[key]; duplicate {
			return fmt.Errorf("duplicate requested slot %s", key)
		}
		expected[key] = struct{}{}
	}
	if err := client.CallContext(ctx, &account, "eth_getProof", address, keys, rpc.BlockNumberOrHashWithHash(anchor, true)); err != nil {
		return fmt.Errorf("get storage proof: %w", err)
	}
	bundle := evidence.Bundle{Version: 1, Header: header, Account: account}
	snapshot, err := bundle.Verify(anchor, address)
	if err != nil {
		return err
	}
	if len(snapshot.Slots) != len(expected) {
		return fmt.Errorf("RPC proof does not cover the requested slots")
	}
	for key := range expected {
		if _, found := snapshot.Slots[key]; !found {
			return fmt.Errorf("RPC proof is missing requested slot %s", key)
		}
	}
	if defaultSlots {
		if _, err := proxyValues(snapshot); err != nil {
			return err
		}
	}
	return writeJSON(out, bundle)
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
		if common.BytesToHash(word[12:]) != word {
			return nil, fmt.Errorf("%s slot does not encode an address", slot.name)
		}
		values = append(values, common.BytesToAddress(word[12:]))
	}
	return values, nil
}

func writeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
