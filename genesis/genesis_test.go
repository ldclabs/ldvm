// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"math/big"
	"os"
	"testing"

	"github.com/ldclabs/ldvm/ld"
)

func TestGenesis(t *testing.T) {
	address1, _ := ld.EthIDFromString("0xa54701B7b7a8f2E9545b4bB90465a0f45C82A84B")
	address2, _ := ld.EthIDFromString("0x3Fb2B2BEBf856C523aA36637e823612a2cB3EEa9")

	file := "./genesis.json"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Read %s failed: %v", file, err)
	}

	gs, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if gs.ChainConfig.ChainID != uint64(2357) ||
		gs.ChainConfig.MaxTotalSupply.Cmp(big.NewInt(1000000000000000000)) != 0 ||
		gs.Alloc[address1].Guardians[0] != address2 {
		t.Fatalf("parse genesis failed")
	}
}
