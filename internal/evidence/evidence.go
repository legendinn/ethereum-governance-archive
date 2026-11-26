// Package evidence verifies Ethereum storage snapshots against pinned block hashes.
package evidence

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

type StorageProof struct {
	Key   common.Hash     `json:"key"`
	Value *hexutil.Big    `json:"value"`
	Proof []hexutil.Bytes `json:"proof"`
}

type AccountProof struct {
	Address common.Address  `json:"address"`
	Nodes   []hexutil.Bytes `json:"accountProof"`
	Storage []StorageProof  `json:"storageProof"`
}

type Bundle struct {
	Version int           `json:"version"`
	Header  *types.Header `json:"header"`
	Account AccountProof  `json:"account"`
}

type Snapshot struct {
	BlockHash common.Hash
	Address   common.Address
	Slots     map[common.Hash]common.Hash
}

// Verify requires an independently trusted block hash and account address.
func (b Bundle) Verify(anchor common.Hash, address common.Address) (Snapshot, error) {
	if b.Version != 1 || b.Header == nil || b.Header.Number == nil || b.Header.Difficulty == nil {
		return Snapshot{}, fmt.Errorf("unsupported or incomplete bundle")
	}
	if b.Header.Hash() != anchor {
		return Snapshot{}, fmt.Errorf("header does not match trusted block hash")
	}
	if b.Account.Address != address {
		return Snapshot{}, fmt.Errorf("bundle does not match expected account")
	}
	accountRLP, err := prove(b.Header.Root, address[:], b.Account.Nodes)
	if err != nil {
		return Snapshot{}, fmt.Errorf("account proof: %w", err)
	}
	if len(accountRLP) == 0 {
		return Snapshot{}, fmt.Errorf("account does not exist at anchored block")
	}
	var account types.StateAccount
	if err := rlp.DecodeBytes(accountRLP, &account); err != nil {
		return Snapshot{}, fmt.Errorf("decode account: %w", err)
	}
	if len(b.Account.Storage) == 0 {
		return Snapshot{}, fmt.Errorf("bundle contains no storage proofs")
	}

	snapshot := Snapshot{BlockHash: anchor, Address: address, Slots: make(map[common.Hash]common.Hash)}
	for _, slot := range b.Account.Storage {
		if _, exists := snapshot.Slots[slot.Key]; exists {
			return Snapshot{}, fmt.Errorf("duplicate slot %s", slot.Key)
		}
		raw, err := prove(account.Root, slot.Key[:], slot.Proof)
		if err != nil {
			return Snapshot{}, fmt.Errorf("slot %s: %w", slot.Key, err)
		}
		var value []byte
		if len(raw) != 0 {
			if err := rlp.DecodeBytes(raw, &value); err != nil {
				return Snapshot{}, fmt.Errorf("decode slot %s: %w", slot.Key, err)
			}
			if len(value) == 0 || len(value) > common.HashLength || value[0] == 0 {
				return Snapshot{}, fmt.Errorf("invalid encoded value for slot %s", slot.Key)
			}
		}
		if slot.Value == nil || slot.Value.ToInt().Sign() < 0 || !bytes.Equal(value, slot.Value.ToInt().Bytes()) {
			return Snapshot{}, fmt.Errorf("claimed value differs from proof for slot %s", slot.Key)
		}
		snapshot.Slots[slot.Key] = common.BytesToHash(value)
	}
	return snapshot, nil
}

func prove(root common.Hash, key []byte, nodes []hexutil.Bytes) ([]byte, error) {
	if root == types.EmptyRootHash {
		if len(nodes) != 0 {
			return nil, fmt.Errorf("unexpected proof nodes for empty trie")
		}
		return nil, nil
	}
	db := memorydb.New()
	defer db.Close()
	for _, node := range nodes {
		if err := db.Put(crypto.Keccak256(node), node); err != nil {
			return nil, err
		}
	}
	return trie.VerifyProof(root, crypto.Keccak256(key), db)
}
