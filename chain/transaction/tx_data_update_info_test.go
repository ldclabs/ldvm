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

func TestTxUpdateDataInfo(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateDataInfo{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()

	owner := util.Signer1.Address()
	assert.NoError(err)
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		To:        &approver,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Token:     &constants.NativeToken,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: &util.DataIDEmpty}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data id")

	did := util.DataID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "no thing to update")

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Data:      []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Approver:    &approver,
		ApproveList: []ld.TxType{ld.TypeDeleteData},
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 1028500, got 0")
	bs.CheckoutAccounts()

	ownerAcc := bs.MustAccount(owner)
	ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ not found")
	bs.CheckoutAccounts()

	assert.NoError(bs.SaveData(di))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid version, expected 2, got 1")
	bs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver:    &approver,
		ApproveList: []ld.TxType{ld.TypeDeleteData},
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid signatures for data keepers")
	bs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver: &approver,
		ApproveList: []ld.TxType{
			ld.TypeUpdateDataInfo,
			ld.TypeUpdateDataInfoByAuth,
			ld.TypeDeleteData},
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		SigClaims: &ld.SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    di.ID,
			Audience:   di.ModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.HashFromData(di.Data),
		},
	}
	sig, err := util.Signer2.Sign(input.SigClaims.Bytes())
	assert.NoError(err)
	input.Sig = &sig

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	ownerGas := tt.Gas()
	assert.Equal(ownerGas*bctx.Price,
		itx.(*TxUpdateDataInfo).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateDataInfo).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-ownerGas*(bctx.Price+100),
		ownerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(di.Version+1, di2.Version)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer2.Address()}, di.Keepers)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)
	assert.Equal(di.Data, di2.Data)
	assert.Nil(di.Approver)
	assert.NotNil(di2.Approver)
	assert.Equal(util.Signer2.Address(), *di2.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(ld.TxTypes{
		ld.TypeUpdateDataInfo,
		ld.TypeUpdateDataInfoByAuth,
		ld.TypeDeleteData}, di2.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateDataInfo","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","approveList":["TypeUpdateDataInfo","TypeUpdateDataInfoByAuth","TypeDeleteData"],"sigClaims":{"iss":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","sub":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","aud":"LM111111111111111111116DBWJs","exp":100,"nbf":0,"iat":1,"cti":"bPfPyx1epFu5ges4t55fJYpS4HrQFr5zCPfT3mzNmL8XCDAZP"},"sig":"68446e6a605b4a6ed476a851ece334aa74855158470c64011cd29a08d8e0ac516869194007619c26dee4df3da012d4268f9d863f7a2f1130de540d4b9a24008f00"},"signatures":["01930101bbe6330496dfb924425165eb7c960d897299d089bd3f7172532f22d340105f1eeef266cc1c89eab534ee20bc045a5946c1f76127bebcc98596a9d94900","3a88cf918f603e465257f00e19237634ea37e498654451fb141fc5bb2c99bcc9280fdbe6d19e079c46cd11d53e859bf592826def1f74cc8e750c41542d27b58e01"],"id":"2pBpq3Y2HBobmw6GgNjAmtJtFouHZkkAsCr2HwKvNXqW3dmKzN"}`, string(jsondata))

	input = &ld.TxUpdater{ID: &did, Version: 3,
		Approver: &util.EthIDEmpty,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid signature for data approver")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	ownerGas += tt.Gas()
	di2, err = bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(4), di2.Version)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)
	assert.Equal(di.Data, di2.Data)
	assert.Nil(di2.Approver)
	assert.Nil(di2.ApproveList)

	// clear threshold
	input = &ld.TxUpdater{ID: &did, Version: 4,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	ownerGas += tt.Gas()
	di2, err = bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(5), di2.Version)
	assert.Equal(uint16(0), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)

	// clear keepers
	input = &ld.TxUpdater{ID: &did, Version: 5,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	ownerGas += tt.Gas()
	di2, err = bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(6), di2.Version)
	assert.Equal(uint16(0), di2.Threshold)
	assert.Equal(util.EthIDs{}, di2.Keepers)

	// can't update keepers
	input = &ld.TxUpdater{ID: &did, Version: 6,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid signatures for data keepers")
	bs.CheckoutAccounts()

	// can't update data
	input = &ld.TxUpdater{ID: &did, Version: 6,
		Data: []byte(`421`),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	tt = txData.ToTransaction()
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid signatures for data keepers")
	bs.CheckoutAccounts()

	// can't delete data
	input = &ld.TxUpdater{ID: &did, Version: 6, Data: []byte(`421`)}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid signatures for data keepers")

	assert.NoError(bs.VerifyState())
}
