// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"testing"
)

func TestBuilder(t *testing.T) {
	sch := `
	type SomeModel {String:Any}
	`
	_, err := NewLDBuilder("SomeModel", []byte(sch), nil)
	if err != nil {
		t.Fatalf("NewLDBuilder failed: %v", err)
	}
	// v1 := map[string]interface{}{
	// 	"name":    "tom",
	// 	"age":     18,
	// 	"Follows": []string{"张三"},
	// }

	// bytes, err := ldb.Marshal(&v1)
	// if err != nil {
	// 	t.Fatalf("Marshal failed: %v", err)
	// }

	// p, err := ldb.Unmarshal(bytes)
	// if err != nil {
	// 	t.Fatalf("Unmarshal failed: %v", err)
	// }
	// v2, ok := p.(map[string]interface{})
	// if !ok {
	// 	t.Fatalf("should type ok")
	// }
	// if v2["name"] != v1["name"] || v2["age"] != v1["age"] {
	// 	t.Fatalf("should equal")
	// }
}
