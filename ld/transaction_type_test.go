// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxType(t *testing.T) {
	assert := assert.New(t)

	set := map[uint8]struct{}{}
	for _, ty := range AllTxTypes {
		set[uint8(ty)] = struct{}{}
		assert.NotEqual("UnknownTx", ty.String(), fmt.Sprintf("invalid TxType %d", ty))
		assert.True(ty.Gas() < 10000)
		switch ty {
		case TypeTest:
			assert.Equal(uint8(0), uint8(ty))
			assert.True(AllTxTypes.Has(ty))
		case TypeEth:
			assert.Equal(uint8(1), uint8(ty))
			assert.True(TransferTxTypes.Has(ty))
		case TypeExchange:
			assert.Equal(uint8(5), uint8(ty))
			assert.True(TransferTxTypes.Has(ty))
		case TypePunish:
			assert.Equal(uint8(16), uint8(ty))
			assert.True(AllTxTypes.Has(ty))
		case TypeUpdateModelKeepers:
			assert.Equal(uint8(18), uint8(ty))
			assert.True(ModelTxTypes.Has(ty))
		case TypeCreateData:
			assert.Equal(uint8(19), uint8(ty))
			assert.True(DataTxTypes.Has(ty))
		case TypeDeleteData:
			assert.Equal(uint8(23), uint8(ty))
			assert.True(DataTxTypes.Has(ty))
		case TypeAddNonceTable:
			assert.Equal(uint8(32), uint8(ty))
			assert.True(AccountTxTypes.Has(ty))
		case TypeRepay:
			assert.Equal(uint8(45), uint8(ty))
			assert.True(AccountTxTypes.Has(ty))
		}
	}
	assert.Equal(len(set), len(AllTxTypes), "should no duplicate value")
}
