// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
)

func TestAccount(t *testing.T) {
	address := ids.ShortID{1, 2, 3, 4}
	ac := &Account{
		Nonce:     1,
		Balance:   big.NewInt(0),
		Threshold: 1,
		Keepers:   []ids.ShortID{address},
	}
	data, err := ac.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	ac2 := &Account{}
	err = ac2.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !ac2.Equal(ac) {
		t.Fatalf("should equal")
	}

	ac.Nonce++
	data2, err := ac.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if bytes.Equal(data, data2) {
		t.Fatalf("should not equal")
	}
}
