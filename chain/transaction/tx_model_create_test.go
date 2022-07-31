// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxCreateModel(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateModel{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")

	owner := util.Signer1.Address()

	txData := &ld.TxData{
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeCreateModel,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeCreateModel,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to together with amount")

	txData = &ld.TxData{
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := &ld.ModelInfo{}
	assert.ErrorContains(input.SyntacticVerify(), "ModelInfo.SyntacticVerify error: invalid name")
	ipldm, err := service.ProfileModel()
	assert.NoError(err)
	input = &ld.ModelInfo{
		Name:      ipldm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      ipldm.Schema(),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateModel,
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
		"insufficient NativeLDC balance, expected 4096400, got 0")
	bs.CheckoutAccounts()

	ownerAcc := bs.MustAccount(owner)
	ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal(tt.Gas()*bctx.Price,
		itx.(*TxCreateModel).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tt.Gas()*100,
		itx.(*TxCreateModel).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tt.Gas()*(bctx.Price+100),
		ownerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	mi, err := bs.LoadModel(itx.(*TxCreateModel).input.ID)
	assert.NoError(err)
	assert.Equal(input.Name, mi.Name)
	assert.Equal(input.Data, mi.Data)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"ProfileService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a09747970652050726f66696c655365727669636520737472756374207b0a0909747970652020202020202020496e74202020202020202020202020202872656e616d6520227422290a09096e616d652020202020202020537472696e67202020202020202020202872656e616d6520226e22290a09096465736372697074696f6e20537472696e67202020202020202020202872656e616d6520226422290a0909696d61676520202020202020537472696e67202020202020202020202872656e616d6520226922290a090975726c202020202020202020537472696e67202020202020202020202872656e616d6520227522290a0909666f6c6c6f777320202020205b494432305d202020202020202020202872656e616d652022667322290a09096d656d6265727320202020206f7074696f6e616c205b494432305d202872656e616d6520226d7322290a0909657874656e73696f6e7320205b416e795d20202020202020202020202872656e616d652022657822290a097d0aed7dc765","id":"LM4jk4H6TitAwi4seemgtuVPRwqbdSbzsKn"},"signatures":["98b85349935fc3f58c3eea84bad6b47b77e7e6103282d9de38f5aaff1eec6d73154e739fb6d892d3ae32c88a1f044272bcae533ff3269c1ce13022a475b6e92b01"],"id":"K452VextWSEbApsBDXA34pp8aSn6MRZrdSY78aq5CgvX99GYH"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxCreateModelGenesis(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	owner := util.Signer1.Address()

	nm, err := service.NameModel()
	assert.NoError(err)
	mi := &ld.ModelInfo{
		Name:      nm.Name(),
		Threshold: *ld.Uint16Ptr(1),
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      nm.Schema(),
	}
	assert.NoError(mi.SyntacticVerify())

	tt := &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: bctx.ChainConfig().ChainID,
		From:    owner,
		Data:    ld.MustMarshal(mi),
	}
	assert.NoError(tt.SyntacticVerify())

	itx, err := NewGenesisTx(tt)
	assert.NoError(err)
	assert.NoError(itx.(GenesisTx).ApplyGenesis(bctx, bs))

	assert.Equal(uint64(0), itx.(*TxCreateModel).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), itx.(*TxCreateModel).from.Nonce())

	mi2, err := bs.LoadModel(itx.(*TxCreateModel).input.ID)
	assert.NoError(err)
	assert.Equal(uint16(1), mi2.Threshold)
	assert.Equal(util.EthIDs{owner}, mi2.Keepers)
	assert.Nil(mi2.Approver)
	assert.Equal(mi.Name, mi2.Name)
	assert.Equal(mi.Data, mi2.Data)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"NameService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a0974797065204e616d655365727669636520737472756374207b0a09096e616d6520202020537472696e6720202020202020202872656e616d6520226e22290a09096c696e6b656420206f7074696f6e616c2049443230202872656e616d6520226c22290a09097265636f726473205b537472696e675d2020202020202872656e616d652022727322290a097d0ad7077bb1","id":"LMG7314Qg687h2DRg3WHszkeM9wHsGJVEDM"},"id":"2Fy56G1oPMf5xW1MTCDHtc4DqX2M6td8TjG2d41yWsaZwBzkFy"}`, string(jsondata))

	pm, err := service.ProfileModel()
	assert.NoError(err)
	mi = &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: *ld.Uint16Ptr(1),
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      pm.Schema(),
	}
	assert.NoError(mi.SyntacticVerify())

	tt = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: bctx.ChainConfig().ChainID,
		Nonce:   1,
		From:    owner,
		Data:    ld.MustMarshal(mi),
	}
	assert.NoError(tt.SyntacticVerify())

	itx, err = NewGenesisTx(tt)
	assert.NoError(err)
	assert.NoError(itx.(GenesisTx).ApplyGenesis(bctx, bs))

	assert.Equal(uint64(0), itx.(*TxCreateModel).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(2), itx.(*TxCreateModel).from.Nonce())

	mi2, err = bs.LoadModel(itx.(*TxCreateModel).input.ID)
	assert.NoError(err)
	assert.Equal(uint16(1), mi2.Threshold)
	assert.Equal(util.EthIDs{owner}, mi2.Keepers)
	assert.Nil(mi2.Approver)
	assert.Equal(mi.Name, mi2.Name)
	assert.Equal(mi.Data, mi2.Data)

	jsondata, err = itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"ProfileService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a09747970652050726f66696c655365727669636520737472756374207b0a0909747970652020202020202020496e74202020202020202020202020202872656e616d6520227422290a09096e616d652020202020202020537472696e67202020202020202020202872656e616d6520226e22290a09096465736372697074696f6e20537472696e67202020202020202020202872656e616d6520226422290a0909696d61676520202020202020537472696e67202020202020202020202872656e616d6520226922290a090975726c202020202020202020537472696e67202020202020202020202872656e616d6520227522290a0909666f6c6c6f777320202020205b494432305d202020202020202020202872656e616d652022667322290a09096d656d6265727320202020206f7074696f6e616c205b494432305d202872656e616d6520226d7322290a0909657874656e73696f6e7320205b416e795d20202020202020202020202872656e616d652022657822290a097d0aed7dc765","id":"LM4TUY7qt97DMQAwr89td6t4JXoHvnfTDbn"},"id":"HhUwLXYTdWjXvy4U3Qh9MjaDkU4W7LLH2W7JrJHWiRfYasife"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
