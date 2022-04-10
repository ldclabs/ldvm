// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package app

import (
	"bytes"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ldclabs/ldvm/ld"
)

func TestName(t *testing.T) {
	address := ids.ShortID{1, 2, 3, 4}
	name := NewName(&Name{
		Name:   "lvdm",
		Linked: ld.EthID(address).String(),
	})
	data, err := name.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	name2 := NewName(nil)
	err = name2.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !name2.Equal(name) {
		t.Fatalf("should equal")
	}

	name.Entity.Extra.Set("ipv4", basicnode.NewString("127.0.0.1"))
	data2, err := name.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if bytes.Equal(data, data2) {
		t.Fatalf("should not equal")
	}

	data, err = name.ToJSON()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	jsonStr := `{"extra":{"ipv4":"127.0.0.1"},"extraMID":"","linked":"0x0102030400000000000000000000000000000000","name":"lvdm"}`
	if string(data) != jsonStr {
		t.Fatalf("should equal, expected %s, got %s", jsonStr, string(data))
	}
}
