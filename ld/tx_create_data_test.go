// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
)

func TestDataMeta(t *testing.T) {
	address := ids.ShortID{1, 2, 3, 4}
	dm := &DataMeta{
		ModelID:   ids.ShortID{4, 5, 6, 7, 8},
		Version:   10,
		Threshold: 1,
		Keepers:   []ids.ShortID{address},
		Data:      []byte("testdata"),
	}
	data, err := dm.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	dm2 := &DataMeta{}
	err = dm2.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !dm2.Equal(dm) {
		t.Fatalf("should equal")
	}

	dm.Version++
	data2, err := dm.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if bytes.Equal(data, data2) {
		t.Fatalf("should not equal")
	}
}
