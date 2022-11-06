// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxAccounter(t *testing.T) {
	assert := assert.New(t)

	var tx *TxAccounter
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxAccounter{Amount: big.NewInt(-1)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid amount")

	tx = &TxAccounter{Keepers: &signer.Keys{}}
	assert.ErrorContains(tx.SyntacticVerify(), "nil threshold")
	tx = &TxAccounter{Threshold: Uint16Ptr(1)}
	assert.ErrorContains(tx.SyntacticVerify(), "nil keepers")
	tx = &TxAccounter{Threshold: Uint16Ptr(1), Keepers: &signer.Keys{}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold, expected <= 0, got 1")
	tx = &TxAccounter{Threshold: Uint16Ptr(1), Keepers: &signer.Keys{signer.Key(util.AddressEmpty[:])}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty secp256k1 key")

	tx = &TxAccounter{ApproveList: &TxTypes{TxType(255)}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid TxType TypeUnknown(255) in approveList")

	tx = &TxAccounter{ApproveList: &TxTypes{TypeTransfer, TypeTransfer}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid approveList, duplicate TxType TypeTransfer")

	tx = &TxAccounter{
		Threshold: Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer1.Key()},
		Amount:    big.NewInt(1000),
		Data:      []byte(`42`),
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"duplicate key jbl8fOziScK5i9wCJsxMKle_UvwKxwPH")

	tx = &TxAccounter{
		Threshold: Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Amount:    big.NewInt(1000),
		Data:      []byte(`42`),
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(tx)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG"],"amount":1000,"data":42}`, string(jsondata))

	tx2 := &TxAccounter{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())

	cbordata2 := tx2.Bytes()
	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
