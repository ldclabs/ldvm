// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"testing"

	"github.com/ldclabs/ldvm/util"
)

func TestModelMeta(t *testing.T) {
	address := util.EthID{1, 2, 3, 4}
	mm := &ModelMeta{
		Name:      "ModelMeta",
		Threshold: 1,
		Keepers:   []util.EthID{address},
		Data:      []byte("testdata"),
	}
	data, err := mm.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	mm2 := &ModelMeta{}
	err = mm2.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	// if !mm2.Equal(mm) {
	// 	t.Fatalf("should equal")
	// }

	// mm.Threshold++
	// data2, err := mm.Marshal()
	// if err != nil {
	// 	t.Fatalf("Marshal failed: %v", err)
	// }
	// if bytes.Equal(data, data2) {
	// 	t.Fatalf("should not equal")
	// }
}
