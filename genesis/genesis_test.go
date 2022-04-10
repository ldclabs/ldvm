// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/ldclabs/ldvm/ld"
)

func TestGenesis(t *testing.T) {
	address1, _ := ld.EthIDFromString("0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	address2, _ := ld.EthIDFromString("0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

	file := "./genesis.json"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Read %s failed: %v", file, err)
	}

	gs, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if gs.Chain.ChainID != uint64(2357) ||
		gs.Chain.MaxTotalSupply.Cmp(big.NewInt(1000000000000000000)) != 0 ||
		gs.Alloc[address1].Keepers[1] != address2 {
		t.Fatalf("parse genesis failed")
	}
	blk, err := gs.ToBlock()
	if err != nil {
		t.Fatalf("ToBlock failed: %v", err)
	}

	data, err = json.Marshal(blk)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	fmt.Printf("\n%s\n", string(data))
	fmt.Println("LLLL", len(blk.Txs[0].Bytes()), len(blk.Txs[1].Bytes()), len(blk.Bytes()))
	t.Fatalf("finish")
}
