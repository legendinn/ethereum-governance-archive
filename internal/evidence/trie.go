package evidence

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/trie"
)

// State and storage tries use hashed keys. Transaction and receipt tries use RLP indices.
func proveStateValue(root common.Hash, key []byte, nodes []hexutil.Bytes) ([]byte, error) {
	if root == types.EmptyRootHash {
		if len(nodes) != 0 {
			return nil, fmt.Errorf("unexpected proof nodes for empty trie")
		}
		return nil, nil
	}
	return proveTrieValue(root, crypto.Keccak256(key), nodes)
}

func proveTrieValue(root common.Hash, key []byte, nodes []hexutil.Bytes) ([]byte, error) {
	db := memorydb.New()
	defer db.Close()
	for _, node := range nodes {
		if err := db.Put(crypto.Keccak256(node), node); err != nil {
			return nil, err
		}
	}
	return trie.VerifyProof(root, key, db)
}
