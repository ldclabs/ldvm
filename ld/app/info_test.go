// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package app

import (
	"bytes"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

func TestInfo(t *testing.T) {
	address := ids.ShortID{1, 2, 3, 4}
	info := NewInfo(&Info{
		Name:    "lvdm",
		Follows: []string{ld.EthID(address).String()},
	})
	data, err := info.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	info2 := NewInfo(nil)
	err = info2.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !info2.Equal(info) {
		t.Fatalf("should equal")
	}

	// ac.Addrs["eth"] = ld.EthID(address).String()
	info.Entity.Url = "http://127.0.0.1"
	data2, err := info.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if bytes.Equal(data, data2) {
		t.Fatalf("should not equal")
	}

	data, err = info.ToJSON()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	jsonStr := `{"addrs":{},"extra":{},"extraMID":"","follows":["0x0102030400000000000000000000000000000000"],"image":"","kyc":"","name":"lvdm","url":"http://127.0.0.1"}`
	if string(data) != jsonStr {
		t.Fatalf("should equal, expected %s, got %s", jsonStr, string(data))
	}
}
