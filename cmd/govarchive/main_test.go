package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/legendinn/ethereum-governance-archive/internal/evidence"
)

const (
	bridgeAddress = "0x99c9fc46f92e8a1c0dec1b1747d010903e884be1"
	beforeBlock   = "0x1df0e378e2618cd0a40cbfc1d6faeb2dff3de609a1c198e6a2534aae84b04794"
	afterBlock    = "0x379caf7ef6649d359d71f115b4e6ff3cd549bd05ba48b1e644b6924ed93aa13d"
	beforeBundle  = "../../examples/optimism-bridge/before.json"
	afterBundle   = "../../examples/optimism-bridge/after.json"
)

func TestHistoricalDiff(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"diff", "-address", bridgeAddress, "-block", beforeBlock,
		"-next-block", afterBlock, beforeBundle, afterBundle}, &out, &out)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Changes []struct{ Slot, Before, After string }
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Changes) != 1 {
		t.Fatalf("expected exactly one changed slot, got %s", out.String())
	}
	change := result.Changes[0]
	if change.Slot != "implementation" ||
		change.Before != "0x64b5a5ed26dcb17370ff4d33a8d503f0fbd06cff" ||
		change.After != "0x78972e88ab8bbb517a36caea23b931bab58ad3c6" {
		t.Fatalf("unexpected implementation change: %+v", change)
	}
}

func TestVerifyRejectsTampering(t *testing.T) {
	data, err := os.ReadFile(beforeBundle)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, errorText string
		mutate          func(*evidence.Bundle)
	}{
		{"header", "trusted block hash", func(b *evidence.Bundle) { b.Header.Root[0] ^= 1 }},
		{"account", "expected account", func(b *evidence.Bundle) { b.Account.Address[0] ^= 1 }},
		{"storage value", "claimed value differs", func(b *evidence.Bundle) { b.Account.Storage[0].Value.ToInt().SetInt64(42) }},
		{"proof node", "account proof", func(b *evidence.Bundle) { b.Account.Nodes[0][0] ^= 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var bundle evidence.Bundle
			if err := json.Unmarshal(data, &bundle); err != nil {
				t.Fatal(err)
			}
			tc.mutate(&bundle)
			modified, err := json.Marshal(bundle)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "modified.json")
			if err := os.WriteFile(path, modified, 0600); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			err = run([]string{"verify", "-address", bridgeAddress, "-block", beforeBlock, path}, &out, &out)
			if err == nil || !strings.Contains(err.Error(), tc.errorText) {
				t.Fatalf("expected %q, got %v", tc.errorText, err)
			}
			if out.Len() != 0 {
				t.Fatalf("unexpected output: %s", &out)
			}
		})
	}
}
