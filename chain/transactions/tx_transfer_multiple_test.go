// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxTransferMultiple(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTransferMultiple{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := cs.MustAccount(signer.Signer1.Key().Address())
	assert.NoError(sender.UpdateKeepers(ld.Uint16Ptr(2),
		&signer.Keys{signer.Signer1.Key(), signer.Signer3.Key()}, nil, nil))

	recipient := cs.MustAccount(signer.Signer2.Key().Address())

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender.ID(),
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender.ID(),
		To:        recipient.ID().Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender.ID(),
		Amount:    big.NewInt(100),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), `nil "to" together with amount`)

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender.ID(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender.ID(),
		Data:      ld.MustMarshal(ld.SendOutputs{}),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "empty SendOutputs")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender.ID(),
		Data: ld.MustMarshal(ld.SendOutputs{
			{To: recipient.ID(), Amount: new(big.Int).SetUint64(constants.LDC)},
		}),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid signatures for sender")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer3))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1131900, got 0")
	cs.CheckoutAccounts()

	sender.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1000000000, got 998868100")
	cs.CheckoutAccounts()

	sender.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxTransferMultiple).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxTransferMultiple).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		sender.Balance().Uint64())
	assert.Equal(uint64(1), sender.Nonce())

	assert.Equal(constants.LDC, recipient.Balance().Uint64())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransferMultiple","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":[{"to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000000}]},"sigs":["NiWYr-MRWUJH8O7FEdcbWmARopZZ7Xd8Yx1RBRqGGxUz5QKRclEHKcWBl8b9Xbk3GtONpPfaDSRdRSIdPVICGQEl2ESU","w2mo8HgwDoG6onyGGqJSjDft8vZBhPYLk-Fy-u86Byh_arHhzPofi2Tv2pAkzmjO18K9grSK2frruCSuZFvGBwMoy4E"],"id":"nDmAeN7TbkYrq1lB0NFGdkykpxe0DbtERmZ8HXyKWyW6jsFe"}`, string(jsondata))

	token := ld.MustNewToken("$LDC")
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender.ID(),
		Token:     token.Ptr(),
		Data: ld.MustMarshal(ld.SendOutputs{
			{To: recipient.ID(), Amount: new(big.Int).SetUint64(constants.LDC)},
			{To: signer.Signer3.Key().Address(), Amount: new(big.Int).SetUint64(constants.LDC)},
			{To: signer.Signer4.Key().Address(), Amount: new(big.Int).SetUint64(constants.LDC)},
		}),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer3))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	sender.Add(token, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient $LDC balance, expected 3000000000, got 1000000000")
	cs.CheckoutAccounts()

	sender.Add(token, new(big.Int).SetUint64(constants.LDC*2))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxTransferMultiple).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxTransferMultiple).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		sender.Balance().Uint64())
	assert.Equal(uint64(0), sender.BalanceOf(token).Uint64())
	assert.Equal(uint64(2), sender.Nonce())

	assert.Equal(constants.LDC, recipient.BalanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC, recipient.BalanceOf(token).Uint64())
	assert.Equal(constants.LDC, cs.MustAccount(signer.Signer3.Key().Address()).BalanceOf(token).Uint64())
	assert.Equal(constants.LDC, cs.MustAccount(signer.Signer4.Key().Address()).BalanceOf(token).Uint64())

	jsondata, err = itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransferMultiple","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","token":"$LDC","data":[{"to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000000},{"to":"0x6962DD0564Fb1f8459624e5b7c5dD9A38b2F990d","amount":1000000000},{"to":"0xaf007738116a90d317f7028a77db4Da8aC58F836","amount":1000000000}]},"sigs":["gMGDXMoczMoTc7a75ld-L5KuiLuicZU9GBCwdqmNSIISGeeb_WmT3cSSX2dCCCzomAqziLOoxbInTuXJ09wwMAABsf2p","5odMmW9EnvO8J1XY9aDUhthTGrDlG50GXn1BHx9IpWGEHQtdCUEpXX3d74rGUTEeVULxYBggqlCDqEfXtf1ICs4fIcw"],"id":"_VrJkggU32UvzNIQ7z1jQqXodcsO1xamgLvTODFlrFHxL2lc"}`, string(jsondata))

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      signer.Signer3.Key().Address(),
		Token:     token.Ptr(),
		Data: ld.MustMarshal(ld.SendOutputs{
			{To: recipient.ID(), Amount: new(big.Int).SetUint64(constants.LDC)},
		}),
	}}

	assert.NoError(ltx.SignWith(signer.Signer3))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid signatures for sender", "can not find ed25519 signer automatically")
	cs.CheckoutAccounts()

	// update account keepers for ed25519 signer
	signer3Acc := cs.MustAccount(signer.Signer3.Key().Address())
	input := ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer3.Key()},
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      signer3Acc.ID(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer3))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 796400, got 0")
	cs.CheckoutAccounts()

	signer3Acc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.Equal(uint16(0), signer3Acc.Threshold())
	assert.Equal(signer.Keys{}, signer3Acc.Keepers())
	assert.NoError(itx.Apply(ctx, cs))

	signer3Gas := ltx.Gas()
	assert.Equal((senderGas+signer3Gas)*ctx.Price,
		itx.(*TxUpdateAccountInfo).ldc.Balance().Uint64())
	assert.Equal((senderGas+signer3Gas)*100,
		itx.(*TxUpdateAccountInfo).miner.Balance().Uint64())
	assert.Equal(constants.LDC-signer3Gas*(ctx.Price+100),
		signer3Acc.Balance().Uint64())
	assert.Equal(uint64(1), signer3Acc.Nonce())
	assert.Equal(uint16(1), signer3Acc.Threshold())
	assert.Equal(signer.Keys{signer.Signer3.Key()}, signer3Acc.Keepers())

	// we can spend the token now~
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferMultiple,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      signer3Acc.ID(),
		Token:     token.Ptr(),
		Data: ld.MustMarshal(ld.SendOutputs{
			{To: recipient.ID(), Amount: new(big.Int).SetUint64(constants.LDC)},
		}),
	}}

	assert.NoError(ltx.SignWith(signer.Signer3))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	assert.NoError(itx.Apply(ctx, cs))
	signer3Gas += ltx.Gas()
	assert.Equal((senderGas+signer3Gas)*ctx.Price,
		itx.(*TxTransferMultiple).ldc.Balance().Uint64())
	assert.Equal((senderGas+signer3Gas)*100,
		itx.(*TxTransferMultiple).miner.Balance().Uint64())
	assert.Equal(constants.LDC-signer3Gas*(ctx.Price+100),
		signer3Acc.Balance().Uint64())
	assert.Equal(uint64(0), signer3Acc.BalanceOf(token).Uint64())
	assert.Equal(uint64(2), signer3Acc.Nonce())

	assert.Equal(constants.LDC, recipient.BalanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2, recipient.BalanceOf(token).Uint64())

	jsondata, err = itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransferMultiple","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x6962DD0564Fb1f8459624e5b7c5dD9A38b2F990d","token":"$LDC","data":[{"to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000000}]},"sigs":["ZF7WmAYlK-_p8ukngiZOSEh0SnV9IMTZk_rKTBUMY23QLHge2lUKJB96kKYAW_hEiUWZhtloSK3m4MA_R3sqBM7BIzc"],"id":"LGhzwbexTcGPU5wVI07yiTf9wrk_nt3VEx5uyEX4A_mk2_IJ"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxTransferMultipleGas(t *testing.T) {
	t.Skip()

	assert := assert.New(t)
	token := ld.MustNewToken("$LDC")
	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	from := cs.MustAccount(signer.Signer1.Key().Address())

	so := make(ld.SendOutputs, 0, 1000)
	for i := 0; i < 1000; i++ {
		recipient := signer.NewSigner()
		so = append(so, ld.SendOutput{
			To:     recipient.Key().Address(),
			Amount: big.NewInt(int64(constants.LDC))})
		tx := &ld.Transaction{Tx: ld.TxData{
			Type:      ld.TypeTransferMultiple,
			ChainID:   ctx.ChainConfig().ChainID,
			Nonce:     1,
			GasTip:    100,
			GasFeeCap: ctx.Price,
			From:      from.id,
			Token:     &token,
			Data:      ld.MustMarshal(so),
		}}
		tx.SignWith(signer.Signer1, signer.Signer2)
		assert.NoError(tx.SyntacticVerify())
		gas := tx.Gas()
		fmt.Printf("recipients: %d, gas/recip: %.1f, txSize: %d, totalGas: %d\n",
			i+1, float64(gas)/float64(i+1), len(tx.Bytes()), gas)
	}
	assert.True(false, "should print gas/recip...")
}
