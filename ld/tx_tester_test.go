// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"
	"testing"

	"github.com/fxamacker/cbor/v2"
	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxTester(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("Address", AddressObject.String())
	assert.Equal("Model", ModelObject.String())
	assert.Equal("Data", DataObject.String())
	assert.Equal("UnknownObjectType(9)", ObjectType(9).String())

	var tx *TxTester
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxTester{
		ObjectType: AddressObject,
		Tests:      cborpatch.Patch{},
	}
	assert.ErrorContains(tx.SyntacticVerify(), "empty objectID")

	tx = &TxTester{
		ObjectType: AddressObject,
		ObjectID:   ids.GenesisAccount.String(),
		Tests:      cborpatch.Patch{},
	}
	assert.ErrorContains(tx.SyntacticVerify(), "empty test")

	tx = &TxTester{
		ObjectType: ObjectType(4),
		ObjectID:   ids.GenesisAccount.String(),
		Tests:      cborpatch.Patch{{Op: cborpatch.OpTest, Path: cborpatch.Path{}}},
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid objectType UnknownObjectType(4)")

	// AddressObject
	tx = &TxTester{
		ObjectType: AddressObject,
		ObjectID:   ids.GenesisAccount.String(),
		Tests: cborpatch.Patch{
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("t"), Value: encoding.MustMarshalCBOR(NativeAccount)},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("n"), Value: encoding.MustMarshalCBOR(uint64(1))},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("b"), Value: encoding.MustMarshalCBOR(new(big.Int).SetUint64(unit.LDC))},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("th"), Value: encoding.MustMarshalCBOR(uint64(1))},
		},
	}
	assert.NoError(tx.SyntacticVerify())
	assert.False(tx.maybeTestData())

	data, err := cbor.Diag(tx.Bytes(), nil)
	require.NoError(t, err)
	// fmt.Println(string(data))
	assert.Equal(`{"ot": 0, "ts": [{1: 6, 3: ["t"], 4: 0}, {1: 6, 3: ["n"], 4: 1}, {1: 6, 3: ["b"], 4: 1000000000}, {1: 6, 3: ["th"], 4: 1}], "oid": "0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff"}`, string(data))

	acc := &Account{
		Nonce:      0,
		Balance:    big.NewInt(0),
		Threshold:  0,
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}
	assert.NoError(acc.SyntacticVerify())
	assert.ErrorContains(tx.Test(acc.Bytes()),
		`test operation for path ["n"] failed, expected 1, got 0`)

	acc = &Account{
		Nonce:      1,
		Balance:    new(big.Int).SetUint64(unit.LDC),
		Threshold:  1,
		Keepers:    signer.Keys{signer.Signer1.Key()},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}
	assert.NoError(acc.SyntacticVerify())
	assert.NoError(tx.Test(acc.Bytes()))

	acc.Balance.Add(acc.Balance, big.NewInt(1))
	assert.NoError(acc.SyntacticVerify())
	assert.ErrorContains(tx.Test(acc.Bytes()),
		`test operation for path ["b"] failed, expected 1000000000, got 1000000001`)

	// TODO test LedgerObject

	// ModelObject
	tx = &TxTester{
		ObjectType: ModelObject,
		ObjectID:   CBORModelID.String(),
		Tests: cborpatch.Patch{
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("n"), Value: encoding.MustMarshalCBOR("NameService")},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("th"), Value: encoding.MustMarshalCBOR(uint64(1))},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("kp", 0), Value: encoding.MustMarshalCBOR(signer.Signer1.Key())},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("ap"), Value: encoding.MustMarshalCBOR(nil)},
		},
	}
	assert.NoError(tx.SyntacticVerify())
	assert.False(tx.maybeTestData())

	data, err = cbor.Diag(tx.Bytes(), nil)
	require.NoError(t, err)
	// fmt.Println(string(data))
	assert.Equal(`{"ot": 2, "ts": [{1: 6, 3: ["n"], 4: "NameService"}, {1: 6, 3: ["th"], 4: 1}, {1: 6, 3: ["kp", 0], 4: h'8db97c7cece249c2b98bdc0226cc4c2a57bf52fc'}, {1: 6, 3: ["ap"], 4: null}], "oid": "AAAAAAAAAAAAAAAAAAAAAAAAAAGIYKah"}`, string(data))

	sch := `
	type ID20 bytes
	type NameService struct {
		name    String        (rename "n")
		linked  nullable ID20 (rename "l")
		records [String]      (rename "rs")
	}
`

	mi := &ModelInfo{
		Name:      "NameService",
		Threshold: 0,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Schema:    sch,
	}
	assert.NoError(mi.SyntacticVerify())
	assert.ErrorContains(tx.Test(mi.Bytes()),
		`test operation for path ["th"] failed, expected 1, got 0`)

	mi = &ModelInfo{
		Name:      "NameService",
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Schema:    sch,
	}
	assert.NoError(mi.SyntacticVerify())
	assert.NoError(tx.Test(mi.Bytes()))

	// DataObject
	tx = &TxTester{
		ObjectType: DataObject,
		ObjectID:   ids.DataID{1, 2, 3}.String(),
		Tests: cborpatch.Patch{
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("v"), Value: encoding.MustMarshalCBOR(uint64(1))},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("th"), Value: encoding.MustMarshalCBOR(uint64(1))},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("kp", 0), Value: encoding.MustMarshalCBOR(signer.Signer1.Key())},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("ap"), Value: encoding.MustMarshalCBOR(signer.Signer2.Key())},
			{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("pl"), Value: encoding.MustMarshalCBOR([]byte(`42`))},
		},
	}
	assert.NoError(tx.SyntacticVerify())
	assert.False(tx.maybeTestData())

	di := &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Approver:  signer.Signer2.Key(),
		Payload:   []byte(`42`),
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(tx.Test(di.Bytes()))

	tx.Tests = append(tx.Tests[:len(tx.Tests)-1],
		&cborpatch.Operation{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("pl", "name"), Value: encoding.MustMarshalCBOR("John")},
		&cborpatch.Operation{Op: cborpatch.OpTest, Path: cborpatch.PathMustFrom("pl", "age"), Value: encoding.MustMarshalCBOR(42)},
	)
	assert.NoError(tx.SyntacticVerify())
	assert.True(tx.maybeTestData())

	data, err = cbor.Diag(tx.Bytes(), nil)
	require.NoError(t, err)
	// fmt.Println(string(data))
	assert.Equal(`{"ot": 3, "ts": [{1: 6, 3: ["v"], 4: 1}, {1: 6, 3: ["th"], 4: 1}, {1: 6, 3: ["kp", 0], 4: h'8db97c7cece249c2b98bdc0226cc4c2a57bf52fc'}, {1: 6, 3: ["ap"], 4: h'44171c37ff5d7b7bb8dcad5c81f16284a229e641'}, {1: 6, 3: ["pl", "name"], 4: "John"}, {1: 6, 3: ["pl", "age"], 4: 42}], "oid": "AQIDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAoWLSv"}`, string(data))

	type person struct {
		Name string `cbor:"name" json:"name"`
		Age  uint   `cbor:"age" json:"age"`
	}

	v := &person{Name: "John", Age: 42}
	di = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Approver:  signer.Signer2.Key(),
		Payload:   encoding.MustMarshalCBOR(v),
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(tx.Test(di.Bytes()))

	di = &DataInfo{
		ModelID:   JSONModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Approver:  signer.Signer2.Key(),
		Payload:   MustMarshalJSON(v),
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(tx.Test(di.Bytes()))
}
