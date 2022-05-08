// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
)

func TestBlock(t *testing.T) {
	address := util.EthID{1, 2, 3, 4}

	tx := &Transaction{
		Type: TypeTransfer,
		To:   address,
	}

	_, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	sids := util.EthIDToStakeSymbol(address)
	blk := &Block{
		Parent:        ids.Empty,
		Height:        0,
		Timestamp:     0,
		Gas:           1000,
		GasPrice:      1000,
		GasRebateRate: 100,
		Miner:         sids[0],
		Shares:        sids,
		Txs:           []*Transaction{tx},
	}
	data, err := blk.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	blk2 := &Block{}
	err = blk2.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !blk2.Equal(blk) {
		t.Fatalf("should equal")
	}

	blk.Height++
	data2, err := blk.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if bytes.Equal(data, data2) {
		t.Fatalf("should not equal")
	}
}
