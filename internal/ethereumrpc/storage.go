// Package ethereumrpc collects evidence from an Ethereum execution RPC.
package ethereumrpc

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/legendinn/ethereum-governance-archive/internal/evidence"
)

// CaptureStorage returns a bundle and its verified snapshot for exactly the requested keys.
func CaptureStorage(ctx context.Context, endpoint string, block common.Hash, address common.Address, keys []common.Hash) (evidence.Bundle, evidence.Snapshot, error) {
	expected := make(map[common.Hash]struct{}, len(keys))
	for _, key := range keys {
		if _, duplicate := expected[key]; duplicate {
			return evidence.Bundle{}, evidence.Snapshot{}, fmt.Errorf("duplicate requested slot %s", key)
		}
		expected[key] = struct{}{}
	}

	client, err := rpc.DialContext(ctx, endpoint)
	if err != nil {
		return evidence.Bundle{}, evidence.Snapshot{}, fmt.Errorf("connect to RPC: %w", err)
	}
	defer client.Close()

	bundle := evidence.Bundle{Version: 1}
	if err := client.CallContext(ctx, &bundle.Header, "eth_getBlockByHash", block, false); err != nil {
		return evidence.Bundle{}, evidence.Snapshot{}, fmt.Errorf("get block header: %w", err)
	}
	if bundle.Header == nil || bundle.Header.Hash() != block {
		return evidence.Bundle{}, evidence.Snapshot{}, fmt.Errorf("RPC did not return the requested block")
	}
	selector := rpc.BlockNumberOrHashWithHash(block, true)
	if err := client.CallContext(ctx, &bundle.Account, "eth_getProof", address, keys, selector); err != nil {
		return evidence.Bundle{}, evidence.Snapshot{}, fmt.Errorf("get storage proof: %w", err)
	}
	snapshot, err := bundle.Verify(block, address)
	if err != nil {
		return evidence.Bundle{}, evidence.Snapshot{}, err
	}
	if len(snapshot.Slots) != len(expected) {
		return evidence.Bundle{}, evidence.Snapshot{}, fmt.Errorf("RPC proof does not cover the requested slots")
	}
	for _, key := range keys {
		if _, found := snapshot.Slots[key]; !found {
			return evidence.Bundle{}, evidence.Snapshot{}, fmt.Errorf("RPC proof is missing requested slot %s", key)
		}
	}
	return bundle, snapshot, nil
}
