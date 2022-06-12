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

	for _, ty := range AllTxTypes {
		assert.NotEqual("TypeUnknown", ty.String(), fmt.Sprintf("invalid TxType %d", ty))
		assert.True(ty.Gas() < 10000)
		switch ty {
		case TypeTest:
			assert.Equal(TxType(0), ty)
			assert.True(AllTxTypes.Has(ty))
		case TypeEth:
			assert.Equal(TxType(1), ty)
			assert.True(TransferTxTypes.Has(ty))
		case TypeExchange:
			assert.Equal(TxType(5), ty)
			assert.True(TransferTxTypes.Has(ty))
		case TypePunish:
			assert.Equal(TxType(16), ty)
			assert.True(AllTxTypes.Has(ty))
		case TypeUpdateModelKeepers:
			assert.Equal(TxType(18), ty)
			assert.True(ModelTxTypes.Has(ty))
		case TypeCreateData:
			assert.Equal(TxType(19), ty)
			assert.False(DataTxTypes.Has(ty))
		case TypeDeleteData:
			assert.Equal(TxType(23), ty)
			assert.True(DataTxTypes.Has(ty))
		case TypeAddNonceTable:
			assert.Equal(TxType(32), ty)
			assert.True(AccountTxTypes.Has(ty))
		case TypeRepay:
			assert.Equal(TxType(45), ty)
			assert.True(AccountTxTypes.Has(ty))
		}
	}

	var ts TxTypes
	assert.NoError(ts.CheckDuplicate())
	assert.NoError(TransferTxTypes.CheckDuplicate())
	assert.NoError(ModelTxTypes.CheckDuplicate())
	assert.NoError(DataTxTypes.CheckDuplicate())
	assert.NoError(AccountTxTypes.CheckDuplicate())
	assert.NoError(AllTxTypes.CheckDuplicate())
	assert.NoError(TokenFromTxTypes.CheckDuplicate())
	assert.NoError(TokenToTxTypes.CheckDuplicate())
	assert.NoError(StakeFromTxTypes0.CheckDuplicate())
	assert.NoError(StakeFromTxTypes1.CheckDuplicate())
	assert.NoError(StakeFromTxTypes2.CheckDuplicate())
	assert.NoError(StakeToTxTypes.CheckDuplicate())

	ts = append(TxTypes{TypeEth}, AllTxTypes...)
	assert.ErrorContains(ts.CheckDuplicate(), "duplicate TxType TypeEth")
}
