package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/legendinn/ethereum-governance-archive/internal/ethereumrpc"
)

func runCapture(args []string, out, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("capture", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	var target snapshotFlags
	target.register(flags)
	endpoint := flags.String("rpc", "", "Ethereum RPC URL (defaults to ETH_RPC_URL)")
	var keys []common.Hash
	flags.Func("slot", "32-byte storage key to capture (repeatable, defaults to ERC-1967 slots)", func(input string) error {
		var key common.Hash
		if err := key.UnmarshalText([]byte(input)); err != nil {
			return err
		}
		keys = append(keys, key)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return err
	}
	address, block, err := target.parse()
	if err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("capture accepts flags only")
	}
	if *endpoint == "" {
		*endpoint = os.Getenv("ETH_RPC_URL")
	}
	if *endpoint == "" {
		return fmt.Errorf("set ETH_RPC_URL or provide -rpc")
	}

	defaultSlots := len(keys) == 0
	if defaultSlots {
		for _, slot := range proxySlots {
			keys = append(keys, slot.key)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	bundle, snapshot, err := ethereumrpc.CaptureStorage(ctx, *endpoint, block, address, keys)
	if err != nil {
		return err
	}
	if defaultSlots {
		if _, err := proxyValues(snapshot); err != nil {
			return err
		}
	}
	return writeJSON(out, bundle)
}
