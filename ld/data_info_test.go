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

	var di *DataInfo
	assert.ErrorContains(di.SyntacticVerify(), "nil pointer")

	di = &DataInfo{Threshold: 1}
	assert.ErrorContains(di.SyntacticVerify(), "invalid threshold")

	di = &DataInfo{Keepers: util.EthIDs{util.EthIDEmpty}}
	assert.ErrorContains(di.SyntacticVerify(), "invalid keepers, empty address exists")

	di = &DataInfo{Version: 1, Approver: &util.EthIDEmpty}
	assert.ErrorContains(di.SyntacticVerify(), "invalid approver")

	di = &DataInfo{ApproveList: TxTypes{TxType(255)}}
	assert.ErrorContains(di.SyntacticVerify(), "invalid TxType TypeUnknown(255) in approveList")

	di = &DataInfo{ApproveList: TxTypes{TypeTransfer, TypeTransfer}}
	assert.ErrorContains(di.SyntacticVerify(), "invalid approveList, duplicate TxType TypeTransfer")

	di = &DataInfo{
		Version:  1,
		Keepers:  util.EthIDs{util.Signer1.Address()},
		Payload:  []byte(`42`),
		TypedSig: []byte{1, 2, 3},
	}
	assert.ErrorContains(di.SyntacticVerify(), "no sigClaims, typed signature should be nil")

	di = &DataInfo{
		Version:   1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   []byte(`42`),
		SigClaims: &SigClaims{},
		TypedSig:  []byte{1, 2, 3},
	}
	assert.ErrorContains(di.SyntacticVerify(), "invalid typed signature")

	di = &DataInfo{
		Version:   1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   []byte(`42`),
		SigClaims: &SigClaims{Issuer: util.DataID{1, 2, 3, 4}},
		TypedSig:  util.Signature{1, 2, 3}.Typed(),
	}
	assert.ErrorContains(di.SyntacticVerify(),
		"SigClaims.SyntacticVerify error: invalid subject")

	di = &DataInfo{
		Version: 0,
		Payload: []byte(`42`),
	}
	assert.NoError(di.SyntacticVerify())

	di = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer1.Address()},
		Payload:   []byte(`42`),
	}
	assert.ErrorContains(di.SyntacticVerify(),
		"invalid keepers, duplicate address 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	di = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
	}
	assert.NoError(di.SyntacticVerify())

	cbordata, err := di.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(di)
	// fmt.Println(string(jsondata))
	assert.NoError(err)
	assert.Equal(`{"mid":"111111111111111111116DBWJs","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"],"payload":42,"id":"11111111111111111111111111111111LpoYY"}`, string(jsondata))

	di2 := &DataInfo{}
	assert.NoError(di2.Unmarshal(cbordata))
	assert.NoError(di2.SyntacticVerify())

	jsondata2, _ := json.Marshal(di2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, di2.Bytes())
	assert.Equal(cbordata, di2.Clone().Bytes())

	di3 := di2.Clone()
	di3.TypedSig = util.Signature{1, 2, 3}.Typed()
	di3.SigClaims = &SigClaims{
		Issuer:     util.DataID{1, 2, 3, 4},
		Subject:    util.DataID{5, 6, 7, 8},
		Expiration: 100,
		IssuedAt:   1,
		CWTID:      util.Hash{9, 10, 11, 12},
	}
	assert.NoError(di3.SyntacticVerify())
	assert.Equal(cbordata, di2.Bytes())
	assert.NotEqual(cbordata, di3.Bytes())
	di3.TypedSig = util.Signer1.MustSign(di3.SigClaims.Bytes())[:]
	jsondata, err = json.Marshal(di3)
	// fmt.Println(string(jsondata))
	assert.NoError(err)
	assert.Equal(`{"mid":"111111111111111111116DBWJs","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"],"payload":42,"sigClaims":{"iss":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","sub":"3DKYW87Qch2qWuSYnU7qRViZ4NJfwPd46XCW2jf3XiiQfKCoE","aud":"111111111111111111116DBWJs","exp":100,"nbf":0,"iat":1,"cti":"4ytusE1c632hPcJTdvDBFCUSde2ENhhsQG4aCNemLWgenkSZA"},"typedSig":"0xef0f0cea3a58f61a17ade4702a6e6262f93928ecbe44eb0b6d23eec4ade2b07a23058f86669a9f191d25df72667b12a75288e95302643bf66d4e82b9735b583201571f1250","id":"11111111111111111111111111111111LpoYY"}`, string(jsondata))

	assert.NoError(di.MarkDeleted(nil))
	assert.Equal(uint64(0), di.Version)
	assert.Nil(di.TypedSig)
	assert.Nil(di.SigClaims)
	assert.Nil(di.Payload)

	cbordata, err = di.Marshal()
	assert.NoError(err)
	di2 = &DataInfo{}
	assert.NoError(di2.Unmarshal(cbordata))
	assert.NoError(di2.SyntacticVerify())
	assert.Equal(cbordata, di2.Bytes())

	assert.NoError(di2.MarkDeleted([]byte(`"test"`)))
	assert.Equal([]byte(`"test"`), []byte(di2.Payload))
}

func TestDataInfoValidSigClaims(t *testing.T) {
	assert := assert.New(t)

	di := &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(di.ValidSigClaims())

	signer, err := di.Signer()
	assert.ErrorContains(err, "invalid typed signature length, expected 66, got 0")
	assert.Equal(util.EthIDEmpty, signer)

	di = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
		TypedSig:  util.Signature{1, 2, 3}.Typed(),
		SigClaims: &SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.Hash{9, 10, 11, 12},
		},
	}
	assert.NoError(di.SyntacticVerify())
	assert.ErrorContains(di.ValidSigClaims(),
		"DataInfo.ValidSigClaims error: invalid data id")

	di = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   []byte(`42`),
		TypedSig:  util.Signature{1, 2, 3}.Typed(),
		SigClaims: &SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.Hash{9, 10, 11, 12},
		},
		ID: util.DataID{5, 6, 7, 8},
	}
	assert.NoError(di.SyntacticVerify())
	assert.ErrorContains(di.ValidSigClaims(),
		"invalid audience, expected 1111111111111111111Ax1asG, got 111111111111111111116DBWJs")

	di = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Payload:   util.MustMarshalCBOR(42),
		TypedSig:  util.Signature{1, 2, 3}.Typed(),
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
	assert.NoError(di.SyntacticVerify())
	assert.ErrorContains(di.ValidSigClaims(),
		"invalid CWT id")

	di.SigClaims.CWTID = util.HashFromData(di.Payload)
	assert.NoError(di.SyntacticVerify())
	assert.NoError(di.ValidSigClaims())

	di.TypedSig = di.TypedSig[1:]
	_, err = di.Signer()
	assert.ErrorContains(err, "invalid typed signature length, expected 66, got 65")

	di.TypedSig = util.Signature{1, 2, 3}.Typed()
	di.TypedSig[0] = 1
	_, err = di.Signer()
	assert.ErrorContains(err, "unknown signature type, expected 0, got 1")

	di.TypedSig = util.Signature{1, 2, 3}.Typed()
	signer, err = di.Signer()
	assert.ErrorContains(err,
		"DataInfo.Signer error: DeriveSigner error: recovery failed")
	assert.Equal(util.EthIDEmpty, signer)

	sig, err := util.Signer1.Sign(di.SigClaims.Bytes())
	assert.NoError(err)
	di.TypedSig = sig.Typed()
	assert.NoError(di.SyntacticVerify())
	assert.NoError(di.ValidSigClaims())

	signer, err = di.Signer()
	assert.NoError(err)
	assert.Equal(util.Signer1.Address(), signer)

	di2 := di.Clone()
	assert.NoError(di2.SyntacticVerify())
	assert.NoError(di2.ValidSigClaims())

	signer, err = di2.Signer()
	assert.NoError(err)
	assert.Equal(util.Signer1.Address(), signer)

	di3 := &DataInfo{}
	assert.NoError(di3.Unmarshal(di.Bytes()))
	assert.NoError(di3.SyntacticVerify())

	assert.ErrorContains(di3.ValidSigClaims(),
		"DataInfo.ValidSigClaims error: invalid data id")

	di3.ID = di.ID
	assert.NoError(di3.ValidSigClaims())
	signer, err = di3.Signer()
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
