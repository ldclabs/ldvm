// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
)

func TestTxUpdateAccountGuardians(t *testing.T) {
	address := ids.ShortID{1, 2, 3, 4}
	tx := &TxUpdateAccountGuardians{
		Threshold: 1,
		Guardians: []ids.ShortID{address},
	}
	data, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	tx2 := &TxUpdateAccountGuardians{}
	err = tx2.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !tx2.Equal(tx) {
		t.Fatalf("should equal")
	}

	tx.Threshold++
	data2, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if bytes.Equal(data, data2) {
		t.Fatalf("should not equal")
	}
}
