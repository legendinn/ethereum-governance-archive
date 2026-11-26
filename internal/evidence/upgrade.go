package evidence

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

// UpgradeProof binds a transaction and its receipt to the same index in a block.
type UpgradeProof struct {
	Version      int             `json:"version"`
	Header       *types.Header   `json:"header"`
	Index        *uint64         `json:"index"`
	TxNodes      []hexutil.Bytes `json:"transactionProof"`
	ReceiptNodes []hexutil.Bytes `json:"receiptProof"`
}

func (p UpgradeProof) Verify(block, transaction common.Hash, proxy, implementation common.Address) error {
	if p.Version != 1 || p.Header == nil || p.Header.Number == nil || p.Header.Difficulty == nil || p.Index == nil {
		return fmt.Errorf("unsupported or incomplete upgrade proof")
	}
	if p.Header.Hash() != block {
		return fmt.Errorf("header does not match trusted block hash")
	}
	// Transaction and receipt trie keys are RLP indices, with no Keccak prehash.
	key := rlp.AppendUint64(nil, *p.Index)
	txBytes, err := proveTrieValue(p.Header.TxHash, key, p.TxNodes)
	if err != nil {
		return fmt.Errorf("transaction proof: %w", err)
	}
	var tx types.Transaction
	if err := tx.UnmarshalBinary(txBytes); err != nil {
		return fmt.Errorf("decode transaction: %w", err)
	}
	if tx.Hash() != transaction {
		return fmt.Errorf("proven transaction does not match expected hash")
	}

	receiptBytes, err := proveTrieValue(p.Header.ReceiptHash, key, p.ReceiptNodes)
	if err != nil {
		return fmt.Errorf("receipt proof: %w", err)
	}
	var receipt types.Receipt
	if err := receipt.UnmarshalBinary(receiptBytes); err != nil {
		return fmt.Errorf("decode receipt: %w", err)
	}
	if receipt.Type != tx.Type() || len(receipt.PostState) != 0 || receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("expected a successful receipt matching the transaction type")
	}

	event := crypto.Keccak256Hash([]byte("Upgraded(address)"))
	argument := common.BytesToHash(implementation[:])
	for _, log := range receipt.Logs {
		if log != nil && log.Address == proxy && len(log.Topics) == 2 &&
			log.Topics[0] == event && log.Topics[1] == argument && len(log.Data) == 0 {
			return nil
		}
	}
	return fmt.Errorf("receipt does not contain the expected Upgraded event")
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
