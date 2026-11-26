package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/legendinn/ethereum-governance-archive/internal/evidence"
)

func TestCompoundUpgradeProof(t *testing.T) {
	const (
		path           = "../../examples/compound-usdc/upgrade-proof.json"
		block          = "0xdd782654955f5b0ffb41ab185b43a92e7220d0c07901fb29ee676577956c5a72"
		tx             = "0xf9df4d7016a592d1135dee34cacae43030c27c3eb69a12281385faa6bd487f56"
		proxy          = "0xc3d688b66703497daa19211eedff47f25384cdc3"
		implementation = "0x2c045e9a1afb125898d6e736d1ff09fa28788e9d"
	)
	var output bytes.Buffer
	err := run([]string{"verify-upgrade", "-block", block, "-tx", tx, "-address", proxy,
		"-implementation", implementation, path}, &output, &output)
	if err != nil || !strings.Contains(output.String(), tx) {
		t.Fatalf("output %q, error %v", output.String(), err)
	}
	for _, tc := range []struct {
		name          string
		expectedError string
		edit          func(*testing.T, *evidence.UpgradeProof)
	}{
		{"different index", "transaction proof", func(_ *testing.T, proof *evidence.UpgradeProof) { *proof.Index += 10000 }},
		{"altered event", "receipt proof", func(t *testing.T, proof *evidence.UpgradeProof) {
			address := common.HexToAddress(implementation)
			for _, node := range proof.ReceiptNodes {
				if start := bytes.Index(node, address[:]); start >= 0 {
					node[start] ^= 1
					return
				}
			}
			t.Fatal("implementation address missing from receipt proof nodes")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var proof evidence.UpgradeProof
			if err := readBundle(path, &proof); err != nil {
				t.Fatal(err)
			}
			tc.edit(t, &proof)
			err := proof.Verify(common.HexToHash(block), common.HexToHash(tx), common.HexToAddress(proxy), common.HexToAddress(implementation))
			if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
				t.Fatalf("expected %q, got %v", tc.expectedError, err)
			}
		})
	}
	var proof evidence.UpgradeProof
	if err := readBundle(path, &proof); err != nil {
		t.Fatal(err)
	}
	err = proof.Verify(common.HexToHash(block), common.HexToHash(tx), common.HexToAddress(proxy), common.Address{})
	if err == nil || !strings.Contains(err.Error(), "expected Upgraded event") {
		t.Fatalf("expected the mismatched implementation to be rejected, got %v", err)
	}
}
