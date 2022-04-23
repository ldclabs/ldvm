// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
)

func TestTransaction(t *testing.T) {
	address := ids.ShortID{1, 2, 3, 4}
	tx := &Transaction{
		Type:    TypeTransfer,
		ChainID: gChainID,
		To:      address,
	}

	if err := tx.SyntacticVerify(); err != nil {
		t.Fatalf("SyntacticVerify failed: %v", err)
	}

	if tx.ID() == ids.Empty {
		t.Fatalf("id should not be Empty")
	}

	data := tx.Bytes()

	tx2 := &Transaction{}
	err := tx2.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !tx2.Equal(tx) {
		t.Fatalf("should equal")
	}

	tx.Gas++
	data2, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if bytes.Equal(data, data2) {
		t.Fatalf("should not equal")
	}
}
