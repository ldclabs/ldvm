// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"
)

// 0xf810f75a70ca722c41bf08500031b511147a1b3be07b481c30d62cff31fc9939
// G4uZHWhjfHoMUANsMZhrW7egB4daNUAV6
const address1 = "0xa54701B7b7a8f2E9545b4bB90465a0f45C82A84B"

// 0xeafebd2e00c325c753cd5d51875fce70bde0db5ae2a0d69f243394b9e0aed488
// 6ooeFpbHdY1BMsvmxjgGiNLdepVnqzG3h
const address2 = "0x3Fb2B2BEBf856C523aA36637e823612a2cB3EEa9"

func TestEthID(t *testing.T) {
	id1, err := EthIDFromString(address1)
	if err != nil {
		t.Fatalf("EthIDFromString(%s) error: %v", address1, err)
	}
	id2, err := EthIDFromString("a54701B7b7a8f2E9545b4bB90465a0f45C82A84B")
	if err != nil {
		t.Fatalf("EthIDFromString(%s) error: %v", address1, err)
	}
	id3, err := EthIDFromString("G4uZHWhjfHoMUANsMZhrW7egB4daNUAV6")
	if err != nil {
		t.Fatalf("EthIDFromString(%s) error: %v", address1, err)
	}
	if id1 != id2 || id1 != id3 {
		t.Fatalf("EthIDFromString error")
	}

	id, _ := EthIDFromString(address2)
	eids := make([]EthID, 0)
	err = json.Unmarshal([]byte(`[
		"0x3Fb2B2BEBf856C523aA36637e823612a2cB3EEa9",
	  "3Fb2B2BEBf856C523aA36637e823612a2cB3EEa9",
	  "6ooeFpbHdY1BMsvmxjgGiNLdepVnqzG3h",
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
	mid := "M7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"
	id, err := ModelIDFromString(mid)
	if err != nil {
		t.Fatalf("ModelIDFromString(%s) error: %v", mid, err)
	}

	mids := make([]ModelID, 0)
	err = json.Unmarshal([]byte(`[
		"M7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
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
	mid := "D7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"
	id, err := DataIDFromString(mid)
	if err != nil {
		t.Fatalf("DataIDFromString(%s) error: %v", mid, err)
	}

	mids := make([]DataID, 0)
	err = json.Unmarshal([]byte(`[
		"D7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
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
