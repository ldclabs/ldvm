// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	cborpatch "github.com/ldclabs/cbor-patch"
	jsonpatch "github.com/ldclabs/json-patch"
	"github.com/ldclabs/ldvm/util"
)

func TestSigClaims(t *testing.T) {
	assert := assert.New(t)

	var sc *SigClaims
	assert.ErrorContains(sc.SyntacticVerify(), "nil pointer")

	sc = &SigClaims{}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid issuer")

	sc = &SigClaims{Issuer: util.DataID{1, 2, 3, 4}}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid subject")

	sc = &SigClaims{Issuer: util.DataID{1, 2, 3, 4}, Subject: util.DataID{5, 6, 7, 8}}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid expiration time")

	sc = &SigClaims{
		Issuer:     util.DataID{1, 2, 3, 4},
		Subject:    util.DataID{5, 6, 7, 8},
		Expiration: 100,
	}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid issued time")

	sc = &SigClaims{
		Issuer:     util.DataID{1, 2, 3, 4},
		Subject:    util.DataID{5, 6, 7, 8},
		Expiration: 100,
		IssuedAt:   1,
	}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid CWT id")

	sc = &SigClaims{
		Issuer:     util.DataID{1, 2, 3, 4},
		Subject:    util.DataID{5, 6, 7, 8},
		Expiration: 100,
		IssuedAt:   1,
		CWTID:      util.Hash{9, 10, 11, 12},
	}
	assert.NoError(sc.SyntacticVerify())

	cbordata, err := sc.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(sc)
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"iss":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","sub":"3DKYW87Qch2qWuSYnU7qRViZ4NJfwPd46XCW2jf3XiiQfKCoE","aud":"111111111111111111116DBWJs","exp":100,"nbf":0,"iat":1,"cti":"4ytusE1c632hPcJTdvDBFCUSde2ENhhsQG4aCNemLWgenkSZA"}`, string(jsondata))

	sc2 := &SigClaims{}
	assert.NoError(sc2.Unmarshal(cbordata))
	assert.NoError(sc2.SyntacticVerify())

	jsondata2, _ := json.Marshal(sc2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, sc2.Bytes())

	sc2.NotBefore = 10
	assert.NoError(sc2.SyntacticVerify())
	assert.NotEqual(cbordata, sc2.Bytes())
}

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
		Payload: []byte(`42`),
		Sig:     &util.Signature{1, 2, 3},
	}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid signature claims")

	tx = &DataInfo{
		Version:   1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   []byte(`42`),
		SigClaims: &SigClaims{},
	}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid signature")

	tx = &DataInfo{
		Version:   1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   []byte(`42`),
		Sig:       &util.Signature{1, 2, 3},
		SigClaims: &SigClaims{Issuer: util.DataID{1, 2, 3, 4}},
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"SigClaims.SyntacticVerify error: invalid subject")

	tx = &DataInfo{
		Version: 0,
		Payload: []byte(`42`),
	}
	assert.NoError(tx.SyntacticVerify())

	tx = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer1.Address()},
		Payload:   []byte(`42`),
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid keepers, duplicate address 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	tx = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
	}
	assert.NoError(tx.SyntacticVerify())

	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	// fmt.Println(string(jsondata))
	assert.NoError(err)
	assert.Equal(`{"mid":"111111111111111111116DBWJs","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"],"payload":42,"id":"11111111111111111111111111111111LpoYY"}`, string(jsondata))

	tx2 := &DataInfo{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())

	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, tx2.Bytes())
	assert.Equal(cbordata, tx2.Clone().Bytes())

	tx3 := tx2.Clone()
	tx3.Sig = &util.Signature{1, 2, 3}
	tx3.SigClaims = &SigClaims{
		Issuer:     util.DataID{1, 2, 3, 4},
		Subject:    util.DataID{5, 6, 7, 8},
		Expiration: 100,
		IssuedAt:   1,
		CWTID:      util.Hash{9, 10, 11, 12},
	}
	assert.NoError(tx3.SyntacticVerify())
	assert.Equal(cbordata, tx2.Bytes())
	assert.NotEqual(cbordata, tx3.Bytes())
	tx3.Sig = util.Signer1.MustSign(tx3.SigClaims.Bytes())
	jsondata, err = json.Marshal(tx3)
	// fmt.Println(string(jsondata))
	assert.NoError(err)
	assert.Equal(`{"mid":"111111111111111111116DBWJs","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"],"payload":42,"sigClaims":{"iss":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","sub":"3DKYW87Qch2qWuSYnU7qRViZ4NJfwPd46XCW2jf3XiiQfKCoE","aud":"111111111111111111116DBWJs","exp":100,"nbf":0,"iat":1,"cti":"4ytusE1c632hPcJTdvDBFCUSde2ENhhsQG4aCNemLWgenkSZA"},"sig":"ef0f0cea3a58f61a17ade4702a6e6262f93928ecbe44eb0b6d23eec4ade2b07a23058f86669a9f191d25df72667b12a75288e95302643bf66d4e82b9735b583201","id":"11111111111111111111111111111111LpoYY"}`, string(jsondata))

	assert.NoError(tx.MarkDeleted(nil))
	assert.Equal(uint64(0), tx.Version)
	assert.Nil(tx.Sig)
	assert.Nil(tx.SigClaims)
	assert.Nil(tx.Payload)

	cbordata, err = tx.Marshal()
	assert.NoError(err)
	tx2 = &DataInfo{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	assert.Equal(cbordata, tx2.Bytes())

	assert.NoError(tx2.MarkDeleted([]byte(`"test"`)))
	assert.Equal([]byte(`"test"`), []byte(tx2.Payload))
}

func TestDataInfoSigner(t *testing.T) {
	assert := assert.New(t)

	tx := &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
	}
	assert.NoError(tx.SyntacticVerify())

	signer, err := tx.Signer()
	assert.ErrorContains(err, "DataInfo.Signer error: invalid signature claims")
	assert.Equal(util.EthIDEmpty, signer)

	tx = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
		Sig:       &util.Signature{1, 2, 3},
	}

	signer, err = tx.Signer()
	assert.ErrorContains(err, "DataInfo.Signer error: invalid signature claims")
	assert.Equal(util.EthIDEmpty, signer)

	tx = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
		Sig:       &util.Signature{1, 2, 3},
		SigClaims: &SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.Hash{9, 10, 11, 12},
		},
	}
	assert.NoError(tx.SyntacticVerify())

	signer, err = tx.Signer()
	assert.ErrorContains(err,
		"DataInfo.Signer error: invalid subject, expected 11111111111111111111111111111111LpoYY, got 3DKYW87Qch2qWuSYnU7qRViZ4NJfwPd46XCW2jf3XiiQfKCoE")
	assert.Equal(util.EthIDEmpty, signer)

	tx = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
		Sig:       &util.Signature{1, 2, 3},
		SigClaims: &SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.Hash{9, 10, 11, 12},
		},
		ID: util.DataID{5, 6, 7, 8},
	}
	assert.NoError(tx.SyntacticVerify())

	signer, err = tx.Signer()
	assert.ErrorContains(err,
		"DataInfo.Signer error: invalid audience, expected 1111111111111111111Ax1asG, got 111111111111111111116DBWJs")
	assert.Equal(util.EthIDEmpty, signer)

	tx = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   util.MustMarshalCBOR(42),
		Sig:       &util.Signature{1, 2, 3},
		SigClaims: &SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Audience:   CBORModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.Hash{9, 10, 11, 12},
		},
		ID: util.DataID{5, 6, 7, 8},
	}
	assert.NoError(tx.SyntacticVerify())

	signer, err = tx.Signer()
	assert.ErrorContains(err,
		"DataInfo.Signer error: invalid CWT id")
	assert.Equal(util.EthIDEmpty, signer)

	tx.SigClaims.CWTID = util.HashFromData(tx.Payload)
	assert.NoError(tx.SyntacticVerify())

	signer, err = tx.Signer()
	assert.ErrorContains(err,
		"DataInfo.Signer error: DeriveSigner error: recovery failed")
	assert.Equal(util.EthIDEmpty, signer)

	sig, err := util.Signer1.Sign(tx.SigClaims.Bytes())
	assert.NoError(err)
	tx.Sig = &sig
	assert.NoError(tx.SyntacticVerify())

	signer, err = tx.Signer()
	assert.NoError(err)
	assert.Equal(util.Signer1.Address(), signer)

	tx2 := tx.Clone()
	assert.NoError(tx2.SyntacticVerify())

	signer, err = tx2.Signer()
	assert.NoError(err)
	assert.Equal(util.Signer1.Address(), signer)

	tx3 := &DataInfo{}
	assert.NoError(tx3.Unmarshal(tx.Bytes()))
	assert.NoError(tx3.SyntacticVerify())

	_, err = tx3.Signer()
	assert.ErrorContains(err,
		"DataInfo.Signer error: invalid subject, expected 11111111111111111111111111111111LpoYY, got 3DKYW87Qch2qWuSYnU7qRViZ4NJfwPd46XCW2jf3XiiQfKCoE")

	tx3.ID = tx.ID
	signer, err = tx3.Signer()
	assert.NoError(err)
	assert.Equal(util.Signer1.Address(), signer)
}

func TestDataInfoPatch(t *testing.T) {
	assert := assert.New(t)

	// with RawModelID

	od := []byte(`42`)
	di := &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   od,
	}

	nd := []byte(`"test"`)
	data, err := di.Patch(nd)
	assert.NoError(err)
	assert.Equal(od, []byte(di.Payload))
	assert.Equal(nd, data)

	type person struct {
		Name string `cbor:"n" json:"name"`
		Age  uint   `cbor:"a" json:"age"`
	}

	v1 := person{Name: "John", Age: 42}

	// with CBORModelID
	od = util.MustMarshalCBOR(v1)
	di = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   od,
	}

	_, err = di.Patch([]byte(`"test"`))
	assert.ErrorContains(err, "invalid CBOR patch")

	cborops := cborpatch.Patch{
		{Op: "replace", Path: "/n", Value: util.MustMarshalCBOR("John X")},
		{Op: "replace", Path: "/a", Value: util.MustMarshalCBOR(uint(18))},
	}
	data, err = di.Patch(util.MustMarshalCBOR(cborops))
	assert.NoError(err)
	assert.Equal(od, []byte(di.Payload))

	v2 := &person{}
	assert.NoError(util.UnmarshalCBOR(data, v2))
	assert.Equal("John X", v2.Name)
	assert.Equal(uint(18), v2.Age)

	// with JSONModelID
	od = MustMarshalJSON(v1)
	di = &DataInfo{
		ModelID:   JSONModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   od,
	}

	_, err = di.Patch([]byte(`"test"`))
	assert.ErrorContains(err, "invalid JSON patch")

	jsonops := jsonpatch.Patch{
		{Op: "replace", Path: "/name", Value: MustMarshalJSON("John X")},
		{Op: "replace", Path: "/age", Value: MustMarshalJSON(uint(18))},
	}
	data, err = di.Patch(MustMarshalJSON(jsonops))
	assert.NoError(err)
	assert.Equal(od, []byte(di.Payload))

	v2 = &person{}
	assert.NoError(json.Unmarshal(data, v2))
	assert.Equal("John X", v2.Name)
	assert.Equal(uint(18), v2.Age)

	// with invalid modelID
	di = &DataInfo{
		ModelID:   util.ModelID{1, 2, 3},
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   od,
	}
	_, err = di.Patch(MustMarshalJSON(jsonops))
	assert.ErrorContains(err,
		"DataInfo.Patch error: unsupport mid 6L5yB2u4uKaHNHEMc4ygsv9c58ZNDTE4")
}
