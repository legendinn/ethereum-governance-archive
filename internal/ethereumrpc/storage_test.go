package ethereumrpc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/legendinn/ethereum-governance-archive/internal/evidence"
)

func TestCaptureStorage(t *testing.T) {
	data, err := os.ReadFile("../../examples/compound-usdc/control/timelock.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture evidence.Bundle
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	block := common.HexToHash("0x5caa47372cefc6b87a7b0e2b18d30b9ab9581ff1f9a427bfc3800f63cb1afae2")
	account := common.HexToAddress("0x6d903f6003cca6255d85cca4d3b5e5146dc33925")
	delaySlot := common.HexToHash("0x2")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     json.RawMessage
			Method string
			Params []json.RawMessage
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		var result any
		switch request.Method {
		case "eth_getBlockByHash":
			result = fixture.Header
		case "eth_getProof":
			var selector rpc.BlockNumberOrHash
			if len(request.Params) != 3 || json.Unmarshal(request.Params[2], &selector) != nil ||
				selector.BlockHash == nil || *selector.BlockHash != block || !selector.RequireCanonical {
				t.Error("expected eth_getProof to use the canonical block hash")
			}
			result = fixture.Account
		default:
			t.Errorf("unexpected method %s", request.Method)
			http.Error(w, "invalid method", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result}); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	_, snapshot, err := CaptureStorage(t.Context(), server.URL, block, account, []common.Hash{{}, delaySlot})
	if err != nil {
		t.Fatal(err)
	}
	if got := snapshot.Slots[delaySlot]; got != common.HexToHash("0x2a300") {
		t.Fatalf("delay slot = %s, want 172800 seconds", got)
	}

	_, _, err = CaptureStorage(t.Context(), server.URL, block, account, []common.Hash{{}, common.HexToHash("0x3")})
	if err == nil || !strings.Contains(err.Error(), "missing requested slot") {
		t.Fatalf("expected a mismatched slot to be rejected, got %v", err)
	}
}
