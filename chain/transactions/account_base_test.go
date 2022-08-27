// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

func TestNativeAccount(t *testing.T) {
	assert := assert.New(t)

	token := ld.MustNewToken("$TEST")
	acc := NewAccount(util.Signer1.Address())
	acc.Init(big.NewInt(0), 2, 2)

	assert.Equal(ld.NativeAccount, acc.Type())
	assert.Equal(true, acc.IsEmpty())
	assert.Equal(true, acc.Valid(ld.NativeAccount))
	assert.Equal(false, acc.Valid(ld.TokenAccount))
	assert.Equal(false, acc.Valid(ld.StakeAccount))
	assert.Equal(uint64(0), acc.Nonce())

	assert.Equal(uint64(0), acc.Balance().Uint64())
	assert.Equal(uint64(0), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), acc.balanceOfAll(token).Uint64())

	assert.NoError(acc.checkBalance(constants.NativeToken, big.NewInt(0)))
	assert.NoError(acc.checkBalance(token, big.NewInt(0)))
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.checkBalance(token, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.checkBalance(token, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, big.NewInt(1)), "insufficient NativeLDC balance")
	assert.ErrorContains(acc.checkBalance(token, big.NewInt(1)), "insufficient $TEST balance, expected 1, got 0")

	for _, ty := range ld.AllTxTypes {
		assert.NoError(acc.CheckAsFrom(ty))
		assert.NoError(acc.CheckAsTo(ty))
	}

	// UpdateKeepers, SatisfySigning, SatisfySigningPlus
	assert.Equal(uint16(0), acc.Threshold())
	assert.Equal(util.EthIDs{}, acc.Keepers())
	assert.True(acc.SatisfySigning(util.EthIDs{util.Signer1.Address()}))
	assert.True(acc.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}))
	assert.False(NewAccount(constants.LDCAccount).SatisfySigning(util.EthIDs{constants.LDCAccount}))
	assert.False(NewAccount(constants.LDCAccount).SatisfySigningPlus(util.EthIDs{constants.LDCAccount}))

	assert.NoError(acc.UpdateKeepers(ld.Uint16Ptr(1), &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, nil, nil))
	assert.True(acc.SatisfySigning(util.EthIDs{util.Signer1.Address()}))
	assert.True(acc.SatisfySigning(util.EthIDs{util.Signer2.Address()}))
	assert.True(acc.SatisfySigning(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}))
	assert.False(acc.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}))
	assert.False(acc.SatisfySigningPlus(util.EthIDs{util.Signer2.Address()}))
	assert.True(acc.SatisfySigningPlus(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}))

	// Add
	assert.ErrorContains(acc.Add(constants.NativeToken, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Add(token, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Add(constants.NativeToken, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.Add(token, big.NewInt(-1)), "invalid amount -1")
	assert.NoError(acc.Add(constants.NativeToken, big.NewInt(100)))
	assert.NoError(acc.Add(token, big.NewInt(100)))
	assert.Equal(uint64(100), acc.Balance().Uint64())
	assert.Equal(uint64(100), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(token).Uint64())
	assert.NoError(acc.Add(constants.NativeToken, big.NewInt(0)))
	assert.NoError(acc.Add(token, big.NewInt(0)))
	assert.Equal(uint64(100), acc.Balance().Uint64())
	assert.Equal(uint64(100), acc.balanceOf(token).Uint64())
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
	assert.Equal(uint64(90), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(token).Uint64())
	assert.NoError(acc.Sub(constants.NativeToken, big.NewInt(0)))
	assert.NoError(acc.Sub(token, big.NewInt(0)))
	assert.Equal(uint64(90), acc.Balance().Uint64())
	assert.Equal(uint64(90), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(token).Uint64())
	assert.ErrorContains(acc.Sub(constants.NativeToken, big.NewInt(100)),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Sub error: insufficient NativeLDC balance, expected 100, got 90")
	assert.ErrorContains(acc.Sub(token, big.NewInt(100)),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Sub error: insufficient $TEST balance, expected 100, got 90")

	// SubByNonce
	assert.ErrorContains(acc.SubByNonce(token, 1, big.NewInt(10)),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).SubByNonce error: invalid nonce, expected 0, got 1")
	assert.NoError(acc.SubByNonce(constants.NativeToken, 0, big.NewInt(10)))
	assert.NoError(acc.SubByNonce(token, 1, big.NewInt(10)))
	assert.Equal(uint64(80), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(80), acc.balanceOfAll(token).Uint64())
	assert.ErrorContains(acc.SubByNonce(constants.NativeToken, 2, big.NewInt(100)),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).SubByNonce error: insufficient NativeLDC balance, expected 100, got 80")
	assert.ErrorContains(acc.SubByNonce(token, 2, big.NewInt(100)),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).SubByNonce error: insufficient $TEST balance, expected 100, got 80")

	// NonceTable
	assert.ErrorContains(acc.SubByNonceTable(token, 12345, 1000, big.NewInt(10)),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).SubByNonceTable error: nonce 1000 not exists at 12345")

	assert.NoError(acc.AddNonceTable(12345, []uint64{1, 2, 3, 4, 0}))
	assert.ErrorContains(acc.AddNonceTable(12345, []uint64{2, 10}),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).AddNonceTable error: nonce 2 exists at 12345")
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 0, big.NewInt(10)))

	assert.ErrorContains(acc.SubByNonceTable(token, 12345, 0, big.NewInt(10)),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).SubByNonceTable error: nonce 0 not exists at 12345")
	assert.ErrorContains(acc.SubByNonceTable(token, 123456, 2, big.NewInt(10)),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).SubByNonceTable error: nonce 2 not exists at 123456")
	assert.NoError(acc.SubByNonceTable(token, 12345, 2, big.NewInt(10)))
	assert.Equal(uint64(70), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(70), acc.balanceOfAll(token).Uint64())

	assert.NoError(acc.AddNonceTable(12345, []uint64{0}))
	assert.Equal([]uint64{0, 1, 3, 4}, acc.ld.NonceTable[12345])
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 1, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 3, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(token, 12345, 0, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(token, 12345, 4, big.NewInt(10)))
	assert.Equal(uint64(50), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(50), acc.balanceOfAll(token).Uint64())
	assert.Equal(0, len(acc.ld.NonceTable))

	for i := uint64(0); i < 1026; i++ {
		assert.NoError(acc.AddNonceTable(i, []uint64{i}))
	}
	assert.Nil(acc.ld.NonceTable[0])
	assert.Nil(acc.ld.NonceTable[1])
	assert.Equal(1024, len(acc.ld.NonceTable))
	assert.ErrorContains(acc.AddNonceTable(100, []uint64{100}),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).AddNonceTable error: too many NonceTable groups, expected <= 1024")

	// Marshal
	data, _, err := acc.Marshal()
	assert.NoError(err)
	acc2, err := ParseAccount(acc.id, data)
	assert.NoError(err)
	assert.Equal(acc.ld.Bytes(), acc2.ld.Bytes())

	// Lending
	cfg := &ld.LendingConfig{
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       big.NewInt(1_000_000),
	}
	assert.NoError(acc.InitLedger(nil))
	assert.NoError(acc.OpenLending(cfg))
	assert.NoError(acc.CloseLending())
}
