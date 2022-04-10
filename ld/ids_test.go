// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"
)

// DvNUrvtQgPynDZN7kFckpjZgmTvW8FX5i
const address1 = "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"

// 7D2dmjrr9Fzg7D6tUQAbPKVdhho4uTmo6
const address2 = "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"

func TestEthID(t *testing.T) {
	id1, err := EthIDFromString(address1)
	if err != nil {
		t.Fatalf("EthIDFromString(%s) error: %v", address1, err)
	}
	id2, err := EthIDFromString("8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	if err != nil {
		t.Fatalf("EthIDFromString(%s) error: %v", address1, err)
	}
	id3, err := EthIDFromString("DvNUrvtQgPynDZN7kFckpjZgmTvW8FX5i")
	if err != nil {
		t.Fatalf("EthIDFromString(%s) error: %v", address1, err)
	}

	if id1 != id2 || id1 != id3 {
		t.Fatalf("EthIDFromString error")
	}

	id, _ := EthIDFromString(address2)
	eids := make([]EthID, 0)
	err = json.Unmarshal([]byte(`[
		"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
	  "44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
	  "7D2dmjrr9Fzg7D6tUQAbPKVdhho4uTmo6",
		"",
		null
	]`), &eids)
	if err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if len(eids) != 5 || id != eids[0] || id != eids[1] || id != eids[2] {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if eids[3] != EthIDEmpty || eids[4] != EthIDEmpty {
		t.Fatalf("Expected EthEmpty")
	}

	if id, err := EthIDFromString(""); err != nil || id != EthIDEmpty {
		t.Fatalf("Expected EthEmpty")
	}
}

func TestModelID(t *testing.T) {
	mid := "LM7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"
	id, err := ModelIDFromString(mid)
	if err != nil {
		t.Fatalf("ModelIDFromString(%s) error: %v", mid, err)
	}

	mids := make([]ModelID, 0)
	err = json.Unmarshal([]byte(`[
		"LM7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
		"",
		null
	]`), &mids)
	if err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if len(mids) != 3 || id != mids[0] {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if mids[1] != ModelIDEmpty || mids[2] != ModelIDEmpty {
		t.Fatalf("Expected DataIDEmpty")
	}

	if id, err := ModelIDFromString(""); err != nil || id != ModelIDEmpty {
		t.Fatalf("Expected ModelIDEmpty")
	}

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if ModelIDFromData(data) != ModelIDFromData(data[:5], data[5:]) {
		t.Fatalf("Expected equal")
	}
}

func TestDataID(t *testing.T) {
	mid := "LD7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"
	id, err := DataIDFromString(mid)
	if err != nil {
		t.Fatalf("DataIDFromString(%s) error: %v", mid, err)
	}

	mids := make([]DataID, 0)
	err = json.Unmarshal([]byte(`[
		"LD7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
		"",
		null
	]`), &mids)
	if err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if len(mids) != 3 || id != mids[0] {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if mids[1] != DataIDEmpty || mids[2] != DataIDEmpty {
		t.Fatalf("Expected DataIDEmpty")
	}

	if id, err := DataIDFromString(""); err != nil || id != DataIDEmpty {
		t.Fatalf("Expected DataIDEmpty")
	}

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if DataIDFromData(data) != DataIDFromData(data[:5], data[5:]) {
		t.Fatalf("Expected equal")
	}
}
