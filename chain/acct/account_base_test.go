// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package acct

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/signer"
)

func TestNativeAccount(t *testing.T) {
	assert := assert.New(t)

	token := ld.MustNewToken("$TEST")
	acc := NewAccount(signer.Signer1.Key().Address())
	acc.Init(big.NewInt(0), big.NewInt(0), 2, 2)

	assert.Equal(ld.NativeAccount, acc.Type())
	assert.Equal(true, acc.IsEmpty())
	assert.Equal(true, acc.Valid(ld.NativeAccount))
	assert.Equal(false, acc.Valid(ld.TokenAccount))
	assert.Equal(false, acc.Valid(ld.StakeAccount))
	assert.Equal(uint64(0), acc.Nonce())

	assert.Equal(uint64(0), acc.Balance().Uint64())
	assert.Equal(uint64(0), acc.balanceOf(token, false).Uint64())
	assert.Equal(uint64(0), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), acc.balanceOfAll(token).Uint64())

	assert.NoError(acc.checkBalance(constants.NativeToken, big.NewInt(0), false))
	assert.NoError(acc.checkBalance(token, big.NewInt(0), false))
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, nil, false), "invalid amount <nil>")
	assert.ErrorContains(acc.checkBalance(token, nil, false), "invalid amount <nil>")
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, big.NewInt(-1), false), "invalid amount -1")
	assert.ErrorContains(acc.checkBalance(token, big.NewInt(-1), false), "invalid amount -1")
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, big.NewInt(1), false), "insufficient NativeLDC balance")
	assert.ErrorContains(acc.checkBalance(token, big.NewInt(1), false), "insufficient $TEST balance, expected 1, got 0")

	for _, ty := range ld.AllTxTypes {
		assert.NoError(acc.ld.CheckAsFrom(ty))
		assert.NoError(acc.ld.CheckAsTo(ty))
	}

	// UpdateKeepers
	assert.Equal(uint16(0), acc.Threshold())
	assert.Equal(signer.Keys{}, acc.Keepers())
	assert.NoError(acc.UpdateKeepers(ld.Uint16Ptr(1), &signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()}, nil, nil))

	// Add
	assert.ErrorContains(acc.Add(constants.NativeToken, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Add(token, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Add(constants.NativeToken, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.Add(token, big.NewInt(-1)), "invalid amount -1")
	assert.NoError(acc.Add(constants.NativeToken, big.NewInt(100)))
	assert.NoError(acc.Add(token, big.NewInt(100)))
	assert.Equal(uint64(100), acc.Balance().Uint64())
	assert.Equal(uint64(100), acc.balanceOf(token, false).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(token).Uint64())
	assert.NoError(acc.Add(constants.NativeToken, big.NewInt(0)))
	assert.NoError(acc.Add(token, big.NewInt(0)))
	assert.Equal(uint64(100), acc.Balance().Uint64())
	assert.Equal(uint64(100), acc.balanceOf(token, false).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(token).Uint64())

	// Sub
	assert.ErrorContains(acc.Sub(constants.NativeToken, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Sub(token, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Sub(constants.NativeToken, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.Sub(token, big.NewInt(-1)), "invalid amount -1")
	assert.NoError(acc.Sub(constants.NativeToken, big.NewInt(10)))
	assert.NoError(acc.Sub(token, big.NewInt(10)))
	assert.Equal(uint64(90), acc.Balance().Uint64())
	assert.Equal(uint64(90), acc.balanceOf(token, false).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(token).Uint64())
	assert.NoError(acc.Sub(constants.NativeToken, big.NewInt(0)))
	assert.NoError(acc.Sub(token, big.NewInt(0)))
	assert.Equal(uint64(90), acc.Balance().Uint64())
	assert.Equal(uint64(90), acc.balanceOf(token, false).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(token).Uint64())
	assert.ErrorContains(acc.Sub(constants.NativeToken, big.NewInt(100)),
		"insufficient transferable NativeLDC balance, expected 100, got 90")
	assert.ErrorContains(acc.Sub(token, big.NewInt(100)),
		"insufficient transferable $TEST balance, expected 100, got 90")

	// SubByNonce
	assert.ErrorContains(acc.SubGasByNonce(token, 1, big.NewInt(10)),
		"invalid nonce, expected 0, got 1")
	assert.NoError(acc.SubGasByNonce(constants.NativeToken, 0, big.NewInt(10)))
	assert.NoError(acc.SubGasByNonce(token, 1, big.NewInt(10)))
	assert.Equal(uint64(80), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(80), acc.balanceOfAll(token).Uint64())
	assert.ErrorContains(acc.SubGasByNonce(constants.NativeToken, 2, big.NewInt(100)),
		"insufficient NativeLDC balance, expected 100, got 80")
	assert.ErrorContains(acc.SubGasByNonce(token, 2, big.NewInt(100)),
		"insufficient $TEST balance, expected 100, got 80")

	// NonceTable
	assert.ErrorContains(acc.SubByNonceTable(token, 12345, 1000, big.NewInt(10)),
		"nonce 1000 not exists at 12345")

	assert.ErrorContains(acc.UpdateNonceTable(1, []uint64{1, 2, 3, 4, 0}),
		"invalid expire, expected >= 2, got 1")
	assert.NoError(acc.UpdateNonceTable(12345, []uint64{1, 2, 3, 4, 0}))
	assert.ErrorContains(acc.UpdateNonceTable(12345, []uint64{1, 2, 3, 4, 2, 10}),
		"duplicate value 2")
	assert.NoError(acc.UpdateNonceTable(12345, []uint64{1, 2, 3, 4, 0, 10}))
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 0, big.NewInt(10)))

	assert.ErrorContains(acc.SubByNonceTable(token, 12345, 0, big.NewInt(10)),
		"nonce 0 not exists at 12345")
	assert.ErrorContains(acc.SubByNonceTable(token, 123456, 2, big.NewInt(10)),
		"nonce 2 not exists at 123456")
	assert.NoError(acc.SubByNonceTable(token, 12345, 2, big.NewInt(10)))
	assert.Equal(uint64(70), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(70), acc.balanceOfAll(token).Uint64())

	assert.NoError(acc.UpdateNonceTable(12345, []uint64{0, 1, 3, 4}))
	assert.Equal([]uint64{0, 1, 3, 4}, acc.ld.NonceTable[12345])
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 1, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 3, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(token, 12345, 0, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(token, 12345, 4, big.NewInt(10)))
	assert.Equal(uint64(50), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(50), acc.balanceOfAll(token).Uint64())
	assert.Equal(0, len(acc.ld.NonceTable))

	for i := uint64(3); i < 1027; i++ {
		assert.NoError(acc.UpdateNonceTable(i, []uint64{i}))
	}
	assert.Nil(acc.ld.NonceTable[2])
	require.NotNil(t, acc.ld.NonceTable[3])
	assert.Equal(1024, len(acc.ld.NonceTable))
	assert.ErrorContains(acc.UpdateNonceTable(1028, []uint64{1028}),
		"too many NonceTable groups, expected <= 1024")

	acc.ld.Timestamp = 28
	assert.NoError(acc.UpdateNonceTable(1028, []uint64{1028}))
	assert.Equal(1000, len(acc.ld.NonceTable))

	// Marshal
	data, _, err := acc.Marshal()
	require.NoError(t, err)
	acc2, err := ParseAccount(acc.ld.ID, data)
	require.NoError(t, err)
	assert.Equal(acc.ld.Bytes(), acc2.ld.Bytes())

	// Lending
	cfg := &ld.LendingConfig{
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       big.NewInt(1_000_000),
	}
	assert.NoError(acc.LoadLedger(false, func() ([]byte, error) { return nil, nil }))
	assert.NoError(acc.OpenLending(cfg))
	assert.NoError(acc.CloseLending())
}
