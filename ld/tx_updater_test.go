// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
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

	tx = &TxUpdater{Keepers: &util.EthIDs{}}
	assert.ErrorContains(tx.SyntacticVerify(), "nil threshold")
	tx2 := &TxUpdater{}
	assert.NoError(tx2.Unmarshal(tx.Bytes()))
	assert.ErrorContains(tx2.SyntacticVerify(), "nil threshold")

	tx = &TxUpdater{Threshold: Uint16Ptr(1)}
	assert.ErrorContains(tx.SyntacticVerify(), "nil keepers")
	tx = &TxUpdater{Threshold: Uint16Ptr(1), Keepers: &util.EthIDs{}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold, expected <= 0, got 1")
	tx = &TxUpdater{Threshold: Uint16Ptr(1), Keepers: &util.EthIDs{util.EthIDEmpty}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid keepers, empty address exists")

	tx = &TxUpdater{ApproveList: TxTypes{TypeCreateData}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid TxType TypeCreateData in approveList")

	tx = &TxUpdater{ApproveList: TxTypes{TypeDeleteData + 1}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid TxType TypeUnknown(24) in approveList")

	tx = &TxUpdater{ApproveList: TxTypes{
		TypeUpdateDataInfo, TypeDeleteData, TypeUpdateDataInfo}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid approveList, duplicate TxType TypeUpdateDataInfo")

	tx = &TxUpdater{SigClaims: &SigClaims{}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"TxUpdater.SyntacticVerify error: nil sig together with sigClaims")

	tx = &TxUpdater{Sig: &util.Signature{}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"TxUpdater.SyntacticVerify error: nil sigClaims together with sig")

	tx = &TxUpdater{
		Sig: &util.Signature{},
		SigClaims: &SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Expiration: 100,
		},
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"TxUpdater.SyntacticVerify error: invalid sigClaims, SigClaims.SyntacticVerify error: invalid issued time")

	tx = &TxUpdater{Token: &util.NativeToken}
	assert.NoError(tx.SyntacticVerify())

	tx = &TxUpdater{
		ID:        &util.DataID{1, 2, 3},
		Version:   1,
		Threshold: Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer1.Address()},
		Approver:  &util.EthIDEmpty,
		Token:     &util.NativeToken,
		To:        &constants.GenesisAccount,
		Amount:    big.NewInt(1000),
		Expire:    uint64(1000),
		Data:      []byte(`"Hello, world!"`),
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid keepers, duplicate address 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	tx = &TxUpdater{
		ID:        &util.DataID{1, 2, 3},
		Version:   1,
		Threshold: Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Approver:  &util.EthIDEmpty,
		Token:     &util.NativeToken,
		To:        &constants.GenesisAccount,
		Amount:    big.NewInt(1000),
		Expire:    uint64(1000),
		Data:      []byte(`"Hello, world!"`),
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"id":"SkB7qHwfMsyF2PgrjhMvtFxJKhuR5ZfVoW9VATWRV4P9jV7J","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"],"approver":"0x0000000000000000000000000000000000000000","token":"","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000,"expire":1000,"data":"Hello, world!"}`, string(jsondata))

	tx2 = &TxUpdater{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, tx2.Bytes())
}
