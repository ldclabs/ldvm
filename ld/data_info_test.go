// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/util"
)

func TestDataInfo(t *testing.T) {
	assert := assert.New(t)

	var tx *DataInfo
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &DataInfo{Threshold: 1}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold")

	tx = &DataInfo{Keepers: util.EthIDs{util.EthIDEmpty}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid keepers, empty address exists")

	tx = &DataInfo{Version: 1, Approver: &util.EthIDEmpty}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid approver")

	tx = &DataInfo{ApproveList: TxTypes{TxType(255)}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid TxType TypeUnknown(255) in approveList")

	tx = &DataInfo{ApproveList: TxTypes{TypeTransfer, TypeTransfer}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid approveList, duplicate TxType TypeTransfer")

	tx = &DataInfo{
		Version: 1,
		Keepers: util.EthIDs{util.Signer1.Address()},
		Data:    []byte(`42`),
		KSig:    util.Signature{1, 2, 3},
	}
	assert.ErrorContains(tx.SyntacticVerify(), "DeriveSigner error: recovery failed")
	tx = &DataInfo{
		Version: 0,
		Data:    []byte(`42`),
	}
	assert.NoError(tx.SyntacticVerify())

	kSig, err := util.Signer1.Sign([]byte(`42`))
	assert.NoError(err)
	tx = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer1.Address()},
		Data:      []byte(`42`),
		KSig:      kSig,
	}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid keepers, duplicate address 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	tx = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Data:      []byte(`42`),
		KSig:      kSig,
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)

	assert.Contains(string(jsondata), `"mid":"LM111111111111111111116DBWJs"`)
	assert.Contains(string(jsondata), `"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"]`)
	assert.Contains(string(jsondata), `"kSig":"`)
	assert.Contains(string(jsondata), `"data":42`)
	assert.NotContains(string(jsondata), `"approver":`)

	tx2 := &DataInfo{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())

	cbordata2 := tx2.Bytes()
	jsondata2, err := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)

	assert.NoError(tx.MarkDeleted(nil))
	assert.Equal(uint64(0), tx.Version)
	assert.Equal(util.SignatureEmpty, tx.KSig)
	assert.Equal(tx2.Data, tx.Data)
	cbordata, err = tx.Marshal()
	assert.NoError(err)
	tx2 = &DataInfo{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 = tx2.Bytes()
	assert.Equal(cbordata, cbordata2)

	assert.NoError(tx2.MarkDeleted([]byte(`"test"`)))
	assert.Equal(uint64(0), tx.Version)
	assert.Equal(util.SignatureEmpty, tx.KSig)
	assert.Equal([]byte(`"test"`), []byte(tx2.Data))
}
