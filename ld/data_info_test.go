// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cborpatch "github.com/ldclabs/cbor-patch"
	jsonpatch "github.com/ldclabs/json-patch"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/encoding"
)

func TestSigClaims(t *testing.T) {
	assert := assert.New(t)

	var sc *SigClaims
	assert.ErrorContains(sc.SyntacticVerify(), "nil pointer")

	sc = &SigClaims{}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid issuer")

	sc = &SigClaims{Issuer: ids.DataID{1, 2, 3, 4}}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid subject")

	sc = &SigClaims{Issuer: ids.DataID{1, 2, 3, 4}, Subject: ids.DataID{5, 6, 7, 8}}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid expiration time")

	sc = &SigClaims{
		Issuer:     ids.DataID{1, 2, 3, 4},
		Subject:    ids.DataID{5, 6, 7, 8},
		Expiration: 100,
	}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid issued time")

	sc = &SigClaims{
		Issuer:     ids.DataID{1, 2, 3, 4},
		Subject:    ids.DataID{5, 6, 7, 8},
		Expiration: 100,
		IssuedAt:   1,
	}
	assert.ErrorContains(sc.SyntacticVerify(), "invalid CWT id")

	sc = &SigClaims{
		Issuer:     ids.DataID{1, 2, 3, 4},
		Subject:    ids.DataID{5, 6, 7, 8},
		Expiration: 100,
		IssuedAt:   1,
		CWTID:      ids.ID32{9, 10, 11, 12},
	}
	assert.NoError(sc.SyntacticVerify())

	cbordata, err := sc.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(sc)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"iss":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","sub":"BQYHCAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADlPJnM","aud":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye","exp":100,"nbf":0,"iat":1,"cti":"CQoLDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARjcYE"}`, string(jsondata))

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

	di = &DataInfo{Keepers: signer.Keys{signer.Key(ids.EmptyAddress[:])}}
	assert.ErrorContains(di.SyntacticVerify(), "empty Secp256k1 key")

	di = &DataInfo{Version: 1, Approver: signer.Key(ids.EmptyAddress[:])}
	assert.ErrorContains(di.SyntacticVerify(), "invalid approver")

	di = &DataInfo{ApproveList: TxTypes{TxType(255)}}
	assert.ErrorContains(di.SyntacticVerify(), "invalid TxType TypeUnknown(255) in approveList")

	di = &DataInfo{ApproveList: TxTypes{TypeTransfer, TypeTransfer}}
	assert.ErrorContains(di.SyntacticVerify(), "invalid approveList, duplicate TxType TypeTransfer")

	sig := make(signer.Sig, 65)
	di = &DataInfo{
		Version: 1,
		Keepers: signer.Keys{signer.Signer1.Key()},
		Payload: []byte(`42`),
		Sig:     &sig,
	}
	assert.ErrorContains(di.SyntacticVerify(), "no sigClaims, signature should be nil")

	di = &DataInfo{
		Version:   1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   []byte(`42`),
		SigClaims: &SigClaims{},
		Sig:       &signer.Sig{},
	}
	assert.ErrorContains(di.SyntacticVerify(), "invalid issuer")

	di = &DataInfo{
		Version:   1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   []byte(`42`),
		SigClaims: &SigClaims{Issuer: ids.DataID{1, 2, 3, 4}},
		Sig:       &sig,
	}
	assert.ErrorContains(di.SyntacticVerify(), "invalid subject")

	di = &DataInfo{
		Version: 0,
		Payload: []byte(`42`),
	}
	assert.NoError(di.SyntacticVerify())

	di = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key(), signer.Signer1.Key()},
		Payload:   []byte(`42`),
	}
	assert.ErrorContains(di.SyntacticVerify(),
		"duplicate key jbl8fOziScK5i9wCJsxMKle_UvwKxwPH")

	di = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Payload:   []byte(`42`),
	}
	assert.NoError(di.SyntacticVerify())

	cbordata, err := di.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(di)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"mid":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye","version":1,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG"],"payload":42,"id":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX"}`, string(jsondata))

	di2 := &DataInfo{}
	assert.NoError(di2.Unmarshal(cbordata))
	assert.NoError(di2.SyntacticVerify())

	jsondata2, _ := json.Marshal(di2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, di2.Bytes())
	assert.Equal(cbordata, di2.Clone().Bytes())

	di3 := di2.Clone()
	di3.SigClaims = &SigClaims{
		Issuer:     ids.DataID{1, 2, 3, 4},
		Subject:    ids.DataID{5, 6, 7, 8},
		Expiration: 100,
		IssuedAt:   1,
		CWTID:      ids.ID32{9, 10, 11, 12},
	}
	di3.Sig = &signer.Sig{1, 2, 3}
	assert.ErrorContains(di3.SyntacticVerify(), "unknown sig AQID_ReApg")
	di3.Sig = signer.Signer1.MustSignData(di3.SigClaims.Bytes()).Ptr()
	assert.NoError(di3.SyntacticVerify())
	assert.Equal(cbordata, di2.Bytes())
	assert.NotEqual(cbordata, di3.Bytes())
	jsondata, err = json.Marshal(di3)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"mid":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye","version":1,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG"],"payload":42,"sigClaims":{"iss":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","sub":"BQYHCAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADlPJnM","aud":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye","exp":100,"nbf":0,"iat":1,"cti":"CQoLDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARjcYE"},"sig":"7w8M6jpY9hoXreRwKm5iYvk5KOy-ROsLbSPuxK3isHojBY-GZpqfGR0l33JmexKnUojpUwJkO_ZtToK5c1tYMgELIV4C","id":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX"}`, string(jsondata))

	assert.NoError(di.MarkDeleted(nil))
	assert.Equal(uint64(0), di.Version)
	assert.Equal(ids.EmptyModelID, di.ModelID)
	assert.Nil(di.Sig)
	assert.Nil(di.SigClaims)
	assert.Nil(di.Payload)

	cbordata, err = di.Marshal()
	require.NoError(t, err)
	di2 = &DataInfo{}
	assert.NoError(di2.Unmarshal(cbordata))
	assert.NoError(di2.SyntacticVerify())
	assert.Equal(cbordata, di2.Bytes())

	assert.NoError(di2.MarkDeleted([]byte(`"test"`)))
	assert.Equal([]byte(`"test"`), []byte(di2.Payload))
}

func TestDataInfoValidSigClaims(t *testing.T) {
	assert := assert.New(t)

	sig := make(signer.Sig, 65)
	di := &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Payload:   []byte(`42`),
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(di.ValidSigClaims())

	di = &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Payload:   []byte(`42`),
		Sig:       &sig,
		SigClaims: &SigClaims{
			Issuer:     ids.DataID{1, 2, 3, 4},
			Subject:    ids.DataID{5, 6, 7, 8},
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      ids.ID32{9, 10, 11, 12},
		},
	}
	assert.NoError(di.SyntacticVerify())
	assert.ErrorContains(di.ValidSigClaims(), "invalid data id")

	di = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Payload:   []byte(`42`),
		Sig:       &sig,
		SigClaims: &SigClaims{
			Issuer:     ids.DataID{1, 2, 3, 4},
			Subject:    ids.DataID{5, 6, 7, 8},
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      ids.ID32{9, 10, 11, 12},
		},
		ID: ids.DataID{5, 6, 7, 8},
	}
	assert.NoError(di.SyntacticVerify())
	assert.ErrorContains(di.ValidSigClaims(),
		"invalid audience, expected AAAAAAAAAAAAAAAAAAAAAAAAAAGIYKah, got AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye")

	di = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Payload:   encoding.MustMarshalCBOR(42),
		Sig:       &sig,
		SigClaims: &SigClaims{
			Issuer:     ids.DataID{1, 2, 3, 4},
			Subject:    ids.DataID{5, 6, 7, 8},
			Audience:   CBORModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      ids.ID32{9, 10, 11, 12},
		},
		ID: ids.DataID{5, 6, 7, 8},
	}
	assert.NoError(di.SyntacticVerify())
	assert.ErrorContains(di.ValidSigClaims(),
		"invalid CWT id")

	di.SigClaims.CWTID = ids.ID32FromData(di.Payload)
	assert.NoError(di.SyntacticVerify())
	assert.NoError(di.ValidSigClaims())
	// TODO

	assert.NoError(di.MarkDeleted(nil))
	assert.Equal(uint64(0), di.Version)
	assert.Equal(CBORModelID, di.ModelID)
	assert.Nil(di.Sig)
	assert.Nil(di.SigClaims)
	assert.Nil(di.Payload)
}

func TestDataInfoPatch(t *testing.T) {
	assert := assert.New(t)

	// with RawModelID

	od := []byte(`42`)
	di := &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   od,
	}

	nd := []byte(`"test"`)
	data, err := di.Patch(nd)
	require.NoError(t, err)
	assert.Equal(od, []byte(di.Payload))
	assert.Equal(nd, data)

	type person struct {
		Name string `cbor:"n" json:"name"`
		Age  uint   `cbor:"a" json:"age"`
	}

	v1 := person{Name: "John", Age: 42}

	// with CBORModelID
	od = encoding.MustMarshalCBOR(v1)
	di = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   od,
	}

	_, err = di.Patch([]byte(`"test"`))
	assert.ErrorContains(err, "invalid CBOR patch")

	cborops := cborpatch.Patch{
		{Op: cborpatch.OpReplace, Path: cborpatch.PathMustFrom("n"), Value: encoding.MustMarshalCBOR("John X")},
		{Op: cborpatch.OpReplace, Path: cborpatch.PathMustFrom("a"), Value: encoding.MustMarshalCBOR(uint(18))},
	}
	data, err = di.Patch(encoding.MustMarshalCBOR(cborops))
	require.NoError(t, err)
	assert.Equal(od, []byte(di.Payload))

	v2 := &person{}
	assert.NoError(encoding.UnmarshalCBOR(data, v2))
	assert.Equal("John X", v2.Name)
	assert.Equal(uint(18), v2.Age)

	// with JSONModelID
	od = MustMarshalJSON(v1)
	di = &DataInfo{
		ModelID:   JSONModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   od,
	}

	_, err = di.Patch([]byte(`"test"`))
	assert.ErrorContains(err, "invalid JSON patch")

	jsonops := jsonpatch.Patch{
		{Op: "replace", Path: "/name", Value: MustMarshalJSON("John X")},
		{Op: "replace", Path: "/age", Value: MustMarshalJSON(uint(18))},
	}
	data, err = di.Patch(MustMarshalJSON(jsonops))
	require.NoError(t, err)
	assert.Equal(od, []byte(di.Payload))

	v2 = &person{}
	assert.NoError(json.Unmarshal(data, v2))
	assert.Equal("John X", v2.Name)
	assert.Equal(uint(18), v2.Age)

	// with invalid modelID
	di = &DataInfo{
		ModelID:   ids.ModelID{1, 2, 3},
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   od,
	}
	_, err = di.Patch(MustMarshalJSON(jsonops))
	assert.ErrorContains(err,
		"unsupport mid AQIDAAAAAAAAAAAAAAAAAAAAAABuT_CC")
}
