// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxUpdater(t *testing.T) {
	assert := assert.New(t)

	var tx *TxUpdater
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxUpdater{Token: &util.TokenSymbol{'a', 'b', 'c'}}
	assert.ErrorContains(tx.SyntacticVerify(),
		`invalid token symbol "0x6162630000000000000000000000000000000000"`)

	tx = &TxUpdater{Amount: big.NewInt(-1)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid amount")

	tx = &TxUpdater{Keepers: &signer.Keys{}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold")
	tx2 := &TxUpdater{}
	assert.NoError(tx2.Unmarshal(tx.Bytes()))
	assert.ErrorContains(tx2.SyntacticVerify(), "invalid threshold")

	tx = &TxUpdater{Threshold: Uint16Ptr(1)}
	assert.ErrorContains(tx.SyntacticVerify(),
		"no keepers, threshold should be nil")
	tx = &TxUpdater{Threshold: Uint16Ptr(1), Keepers: &signer.Keys{}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid threshold, expected <= 0, got 1")
	tx = &TxUpdater{Threshold: Uint16Ptr(1), Keepers: &signer.Keys{signer.Key(util.AddressEmpty[:])}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty Secp256k1 key")

	tx = &TxUpdater{ApproveList: &TxTypes{TypeCreateData}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid TxType TypeCreateData in approveList")

	tx = &TxUpdater{ApproveList: &TxTypes{TypeDeleteData + 1}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid TxType TypeUnknown(25) in approveList")

	tx = &TxUpdater{ApproveList: &TxTypes{
		TypeUpdateDataInfo, TypeDeleteData, TypeUpdateDataInfo}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid approveList, duplicate TxType TypeUpdateDataInfo")

	tx = &TxUpdater{SigClaims: &SigClaims{}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid typed signature")

	tx = &TxUpdater{Sig: &signer.Sig{}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"no sigClaims, typed signature should be nil")

	sig := make(signer.Sig, 65)
	tx = &TxUpdater{
		Sig: &sig,
		SigClaims: &SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Expiration: 100,
		},
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid issued time")

	tx = &TxUpdater{Token: &util.NativeToken}
	assert.NoError(tx.SyntacticVerify())

	tx = &TxUpdater{
		ID:        &util.DataID{1, 2, 3},
		Version:   1,
		Threshold: Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer1.Key()},
		Approver:  &signer.Key{},
		Token:     &util.NativeToken,
		To:        &constants.GenesisAccount,
		Amount:    big.NewInt(1000),
		Expire:    uint64(1000),
		Data:      []byte(`"Hello, world!"`),
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"duplicate key jbl8fOziScK5i9wCJsxMKle_UvwKxwPH")

	tx = &TxUpdater{
		ID:        &util.DataID{1, 2, 3},
		Version:   1,
		Threshold: Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Approver:  &signer.Key{},
		Token:     &util.NativeToken,
		To:        &constants.GenesisAccount,
		Amount:    big.NewInt(1000),
		Expire:    uint64(1000),
		Data:      []byte(`"Hello, world!"`),
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(tx)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"id":"AQIDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAoWLSv","version":1,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG"],"approver":"p__G-A","token":"","to":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","amount":1000,"expire":1000,"data":"Hello, world!"}`, string(jsondata))

	tx2 = &TxUpdater{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, tx2.Bytes())
	assert.NoError(tx2.Approver.ValidOrEmpty())
	assert.ErrorContains(tx2.Approver.Valid(), "empty key")
}
