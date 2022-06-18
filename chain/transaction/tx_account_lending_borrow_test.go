// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxBorrow(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxBorrow{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")

	ldc, err := bs.LoadAccount(constants.LDCAccount)
	assert.NoError(err)
	miner, err := bs.LoadMiner(bctx.Miner())
	assert.NoError(err)
	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	lender, err := bs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to as lender")

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Amount:    new(big.Int).SetUint64(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := &ld.TxTransfer{}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil from as lender")

	input = &ld.TxTransfer{
		From: &lender.id,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	input = &ld.TxTransfer{
		From: &lender.id,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to as borrower")

	input = &ld.TxTransfer{
		From: &lender.id,
		To:   &constants.GenesisAccount,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid from, expected 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF, got 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	input = &ld.TxTransfer{
		From: &lender.id,
		To:   &from.id,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid token, expected NativeLDC, got $LDC")

	input = &ld.TxTransfer{
		From:  &lender.id,
		To:    &from.id,
		Token: &token,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Token:     &constants.NativeToken,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid token, expected $LDC, got NativeLDC")

	input = &ld.TxTransfer{
		From:  &lender.id,
		To:    &from.id,
		Token: &token,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, expected >= 1")

	input = &ld.TxTransfer{
		From:   &lender.id,
		To:     &from.id,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "data expired")

	dueTime := bs.Timestamp()
	dueTimeData, err := ld.MarshalCBOR(dueTime)
	assert.NoError(err)
	input = &ld.TxTransfer{
		From:   &lender.id,
		To:     &from.id,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid dueTime, expected > 1000, got 1000")

	dueTime = bs.Timestamp() + 3600*24
	dueTimeData, err = ld.MarshalCBOR(dueTime)
	assert.NoError(err)
	input = &ld.TxTransfer{
		From:   &lender.id,
		To:     &from.id,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid exSignatures, DeriveSigners: no signature")

	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid gas, expected 694, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 763400, got 0")

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).CheckBorrow failed: invalid lending")

	lcfg := &ld.LendingConfig{
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(lcfg.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      lender.id,
		Data:      ld.MustMarshal(lcfg),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	lenderGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	lender.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(lenderGas*bctx.Price, ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(lenderGas*100, miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lender.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(lender.ld.Lending)
	assert.NotNil(lender.ld.LendingLedger)
	assert.Equal(0, len(lender.ld.LendingLedger))

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"CheckBorrow failed: invalid token, expected NativeLDC, got $LDC")

	input = &ld.TxTransfer{
		From:   &lender.id,
		To:     &from.id,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	fromGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).CheckBorrow failed: insufficient NativeLDC balance, expected 1000000000, got 998549100")

	assert.NoError(lender.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*2)))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).SubByNonceTable failed: nonce 0 not exists at 1000")

	assert.NoError(lender.AddNonceTable(bs.Timestamp(), []uint64{0, 1}))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxBorrow)
	assert.Equal((lenderGas+fromGas)*bctx.Price, ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+fromGas)*100, miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2-fromGas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())
	assert.Equal(1, len(lender.ld.LendingLedger))
	assert.NotNil(lender.ld.LendingLedger[from.id])
	entry := lender.ld.LendingLedger[from.id]
	assert.Equal(constants.LDC, entry.Amount.Uint64())
	assert.Equal(bs.Timestamp(), entry.UpdateAt)
	assert.Equal(bs.Timestamp()+3600*24, entry.DueTime)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeBorrow","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","data":{"from":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","to":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","amount":1000000000,"expire":1000,"data":"0x1a00015568c54d609b"},"signatures":["d23a9c587172465060d4fce7cf73770e3914a2d8f35622856083dab0fdd0bb4a1a8b65671392c6bd15d4eff1ae4cd79516ad5d5ea25571b6278571967c87458701"],"exSignatures":["b82d77943d8761685f7ca432cf5059455a278bb34089cbb02d85b20cd87c430500002e111cb2f693c0c75ba2c96085fce89827c39d9c827b349fb873fac32eaa01"],"gas":646,"id":"2qNonZbDwMaPFV3K8fWhCe27DVerF5WsjDp5cMCKnuzmxwXT6M"}`, string(jsondata))

	// borrow again after a day
	bctx.height++
	bctx.timestamp += 3600 * 24
	bs.height++
	bs.timestamp = bctx.timestamp
	lender.ld.Height++
	lender.ld.Timestamp = bs.Timestamp()
	from.ld.Height++
	from.ld.Timestamp = bs.Timestamp()

	assert.NoError(lender.AddNonceTable(bs.Timestamp(), []uint64{999}))
	input = &ld.TxTransfer{
		Nonce:  999,
		From:   &lender.id,
		To:     &from.id,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp(),
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &lender.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	fromGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(constants.LDC*3-fromGas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lender.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(lender.ld.LendingLedger[from.id])
	entry = lender.ld.LendingLedger[from.id]
	rate := 1 + float64(10_000)/1_000_000
	assert.Equal(uint64(float64(constants.LDC)*rate)+constants.LDC, entry.Amount.Uint64(),
		"with 1 day interest")
	assert.Equal(bs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime, "overwrite to 0")

	assert.NoError(bs.VerifyState())
}
