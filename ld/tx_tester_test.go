// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

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

	ops := TestOps{{}}
	assert.ErrorContains(ops.SyntacticVerify(), "invalid path")

	ops = TestOps{{Path: "/", Value: nil}}
	assert.ErrorContains(ops.SyntacticVerify(), "invalid value")

	var tx *TxTester
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxTester{ObjectType: AddressObject, Tests: TestOps{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty tests")

	tx = &TxTester{ObjectType: ObjectType(4), Tests: TestOps{{Path: "/"}}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"invalid objectType UnknownObjectType(4)")

	tx = &TxTester{ObjectType: AddressObject, Tests: TestOps{{Path: "/"}}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid value")

	// AddressObject
	tx = &TxTester{
		ObjectType: AddressObject,
		ObjectID:   ids.GenesisAccount.String(),
		Tests: TestOps{
			{Path: "/t", Value: encoding.MustMarshalCBOR(NativeAccount)},
			{Path: "/n", Value: encoding.MustMarshalCBOR(uint64(1))},
			{Path: "/b", Value: encoding.MustMarshalCBOR(new(big.Int).SetUint64(unit.LDC))},
			{Path: "/th", Value: encoding.MustMarshalCBOR(uint64(1))},
		},
	}
	assert.NoError(tx.SyntacticVerify())
	assert.False(tx.maybeTestData())

	data, err := json.Marshal(tx)
	require.NoError(t, err)
	// fmt.Println(string(data))
	assert.Equal(`{"objectType":"Address","objectID":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","tests":[{"path":"/t","value":"AF1TRp8"},{"path":"/n","value":"ASdn8Vw"},{"path":"/b","value":"wkQ7msoAEtHq1g"},{"path":"/th","value":"ASdn8Vw"}]}`, string(data))

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
		`test operation for path "/n" failed, expected "1", got "0"`)

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
		`test operation for path "/b" failed, expected "{false [1000000000]}", got "{false [1000000001]}"`)

	// TODO test LedgerObject

	// ModelObject
	tx = &TxTester{
		ObjectType: ModelObject,
		ObjectID:   CBORModelID.String(),
		Tests: TestOps{
			{Path: "/n", Value: encoding.MustMarshalCBOR("NameService")},
			{Path: "/th", Value: encoding.MustMarshalCBOR(uint64(1))},
			{Path: "/kp/0", Value: encoding.MustMarshalCBOR(signer.Signer1.Key())},
			{Path: "/ap", Value: encoding.MustMarshalCBOR(nil)},
		},
	}
	assert.NoError(tx.SyntacticVerify())
	assert.False(tx.maybeTestData())

	data, err = json.Marshal(tx)
	require.NoError(t, err)
	// fmt.Println(string(data))
	assert.Equal(`{"objectType":"Model","objectID":"AAAAAAAAAAAAAAAAAAAAAAAAAAGIYKah","tests":[{"path":"/n","value":"a05hbWVTZXJ2aWNlEFh-6A"},{"path":"/th","value":"ASdn8Vw"},{"path":"/kp/0","value":"VI25fHzs4knCuYvcAibMTCpXv1L8tEv5Hg"},{"path":"/ap","value":"9kV6peQ"}]}`, string(data))

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
		`test operation for path "/th" failed, expected "1", got "0"`)

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
		Tests: TestOps{
			{Path: "/v", Value: encoding.MustMarshalCBOR(uint64(1))},
			{Path: "/th", Value: encoding.MustMarshalCBOR(uint64(1))},
			{Path: "/kp/0", Value: encoding.MustMarshalCBOR(signer.Signer1.Key())},
			{Path: "/ap", Value: encoding.MustMarshalCBOR(signer.Signer2.Key())},
			{Path: "/pl", Value: encoding.MustMarshalCBOR([]byte(`42`))},
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
		TestOp{Path: "/pl/name", Value: encoding.MustMarshalCBOR("John")},
		TestOp{Path: "/pl/age", Value: encoding.MustMarshalCBOR(42)},
	)
	assert.NoError(tx.SyntacticVerify())
	assert.True(tx.maybeTestData())

	data, err = json.Marshal(tx)
	require.NoError(t, err)
	// fmt.Println(string(data))
	assert.Equal(`{"objectType":"Data","objectID":"AQIDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAoWLSv","tests":[{"path":"/v","value":"ASdn8Vw"},{"path":"/th","value":"ASdn8Vw"},{"path":"/kp/0","value":"VI25fHzs4knCuYvcAibMTCpXv1L8tEv5Hg"},{"path":"/ap","value":"VEQXHDf_XXt7uNytXIHxYoSiKeZBrSD8AA"},{"path":"/pl/name","value":"ZEpvaG7CssqR"},{"path":"/pl/age","value":"GCpEY_8t"}]}`, string(data))

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
