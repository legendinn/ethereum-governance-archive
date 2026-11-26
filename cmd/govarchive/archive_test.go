package main

import (
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestArchivedGovernanceEvidence(t *testing.T) {
	const thresholdBefore = "0xaf620161265847641248b219dc681cab3cfd19e7777414c145da3b2e55ecd01d"
	const thresholdAfter = "0x5989cc0cff78aa28e51a5bd9bd80dcae01f39822a93e0004446916245dcacfbb"
	const compoundBefore = "0x5caa47372cefc6b87a7b0e2b18d30b9ab9581ff1f9a427bfc3800f63cb1afae2"
	const compoundAfter = "0xdd782654955f5b0ffb41ab185b43a92e7220d0c07901fb29ee676577956c5a72"
	const compoundProxy = "0xc3d688b66703497daa19211eedff47f25384cdc3"
	const upgradeSafe = "0x5a0aae59d09fccbddb6c6cceb07b7279367c3d2a"
	const foundationSafe = "0x847b5c174615b1b7fdf770882256e2d3e95b9d92"
	const councilSafe = "0xc2819dc788505aac350142a7a707bf9d03e3bd03"
	const compoundAdmin = "0x1ec63b5883c3481134fd50d5daebc83ecd2e8779"
	const timelock = "0x6d903f6003cca6255d85cca4d3b5e5146dc33925"

	cases := []struct {
		path, block, address string
		slots                map[string]string
	}{
		{"optimism-bridge/control/proxy-admin", beforeBlock, "0x543ba4aadbab8f9025686bd03993043599c6fb04", map[string]string{"0x0": upgradeSafe}},
		{"optimism-bridge/control/upgrade-safe", beforeBlock, upgradeSafe, map[string]string{
			"0x3": "0x2", "0x4": "0x2",
			"0xe90b7bceb6e7df5418fb78d8ee546e97c83a08bbccc01a0644d599ccd2a7c2e0": foundationSafe,
			"0x5e7cc3010560e8e197ba9b04a46ffe6535eaa26528be6196218d7efd6419d90a": councilSafe,
			"0x9536f585fb4a85c72bab75c5333dca0c83d3b5944fc1befffa23d431456c9164": "0x1",
		}},
		{"optimism-bridge/control/foundation-safe", beforeBlock, foundationSafe, map[string]string{"0x3": "0x7", "0x4": "0x5"}},
		{"optimism-bridge/control/council-safe", beforeBlock, councilSafe, map[string]string{"0x3": "0xd", "0x4": "0xa"}},
		{"optimism-council-threshold/before", thresholdBefore, councilSafe, map[string]string{"0x3": "0xd", "0x4": "0x4"}},
		{"optimism-council-threshold/after", thresholdAfter, councilSafe, map[string]string{"0x3": "0xd", "0x4": "0xa"}},
		{"optimism-council-threshold/control/bridge", thresholdAfter, bridgeAddress, map[string]string{
			proxySlots[1].key.Hex(): "0x543ba4aadbab8f9025686bd03993043599c6fb04",
		}},
		{"optimism-council-threshold/control/proxy-admin", thresholdAfter, "0x543ba4aadbab8f9025686bd03993043599c6fb04", map[string]string{
			"0x0": upgradeSafe,
		}},
		{"optimism-council-threshold/control/upgrade-safe", thresholdAfter, upgradeSafe, map[string]string{
			"0x3": "0x2", "0x4": "0x2",
			"0xe90b7bceb6e7df5418fb78d8ee546e97c83a08bbccc01a0644d599ccd2a7c2e0": foundationSafe,
			"0x5e7cc3010560e8e197ba9b04a46ffe6535eaa26528be6196218d7efd6419d90a": councilSafe,
			"0x9536f585fb4a85c72bab75c5333dca0c83d3b5944fc1befffa23d431456c9164": "0x1",
		}},
		{"compound-usdc/before", compoundBefore, compoundProxy, map[string]string{
			proxySlots[0].key.Hex(): "0xaec1954467b6d823a9042e9e9d6e4f40111069a9",
			proxySlots[1].key.Hex(): compoundAdmin,
		}},
		{"compound-usdc/after", compoundAfter, compoundProxy, map[string]string{
			proxySlots[0].key.Hex(): "0x2c045e9a1afb125898d6e736d1ff09fa28788e9d",
			proxySlots[1].key.Hex(): compoundAdmin,
		}},
		{"compound-usdc/control/proxy-admin", compoundBefore, compoundAdmin, map[string]string{"0x0": timelock}},
		{"compound-usdc/control/timelock", compoundBefore, timelock, map[string]string{
			"0x0": "0x309a862bbc1a00e45506cb8a802d1ff10004c8c0", "0x2": "0x2a300",
		}},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			snapshot, err := readSnapshot(filepath.Join("..", "..", "examples", tc.path+".json"),
				common.HexToHash(tc.block), common.HexToAddress(tc.address))
			if err != nil {
				t.Fatal(err)
			}
			for slot, expected := range tc.slots {
				value, ok := snapshot.Slots[common.HexToHash(slot)]
				if !ok || value != common.HexToHash(expected) {
					t.Errorf("slot %s = %s (present: %t), expected %s", slot, value, ok, expected)
				}
			}
		})
	}
}
