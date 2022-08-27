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

func TestTxTester(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("Address", AddressObject.String())
	assert.Equal("Model", ModelObject.String())
	assert.Equal("Data", DataObject.String())
	assert.Equal("UnknownObjectType(9)", ObjectType(9).String())

	ops := TestOps{{}}
	assert.ErrorContains(ops.SyntacticVerify(),
		"TestOps.SyntacticVerify error: invalid path")

	ops = TestOps{{Path: "/", Value: nil}}
	assert.ErrorContains(ops.SyntacticVerify(),
		"TestOps.SyntacticVerify error: invalid value")

	var tx *TxTester
	assert.ErrorContains(tx.SyntacticVerify(),
		"TxTester.SyntacticVerify error: nil pointer")

	tx = &TxTester{ObjectType: AddressObject, Tests: TestOps{}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"TxTester.SyntacticVerify error: empty tests")

	tx = &TxTester{ObjectType: ObjectType(4), Tests: TestOps{{Path: "/"}}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"TxTester.SyntacticVerify error: invalid objectType UnknownObjectType(4)")

	tx = &TxTester{ObjectType: AddressObject, Tests: TestOps{{Path: "/"}}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"TxTester.SyntacticVerify error: TestOps.SyntacticVerify error: invalid value")

	// AddressObject
	tx = &TxTester{
		ObjectType: AddressObject,
		ObjectID:   constants.GenesisAccount.String(),
		Tests: TestOps{
			{Path: "/t", Value: util.MustMarshalCBOR(NativeAccount)},
			{Path: "/n", Value: util.MustMarshalCBOR(uint64(1))},
			{Path: "/b", Value: util.MustMarshalCBOR(new(big.Int).SetUint64(constants.LDC))},
			{Path: "/th", Value: util.MustMarshalCBOR(uint64(1))},
		},
	}
	assert.NoError(tx.SyntacticVerify())
	assert.False(tx.maybeTestData())

	data, err := json.Marshal(tx)
	assert.NoError(err)
	// fmt.Println(string(data))
	assert.Equal(`{"objectType":"Address","objectID":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","tests":[{"path":"/t","value":"0x0017afa01d"},{"path":"/n","value":"0x017785459a"},{"path":"/b","value":"0xc2443b9aca00dfb73dae"},{"path":"/th","value":"0x017785459a"}]}`, string(data))

	acc := &Account{
		Nonce:      0,
		Balance:    big.NewInt(0),
		Threshold:  0,
		Keepers:    util.EthIDs{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}
	assert.NoError(acc.SyntacticVerify())
	assert.ErrorContains(tx.Test(acc.Bytes()),
		`TxTester.Test error: test operation for path "/n" failed, expected "1", got "0"`)

	acc = &Account{
		Nonce:      1,
		Balance:    new(big.Int).SetUint64(constants.LDC),
		Threshold:  1,
		Keepers:    util.EthIDs{util.Signer1.Address()},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}
	assert.NoError(acc.SyntacticVerify())
	assert.NoError(tx.Test(acc.Bytes()))

	acc.Balance.Add(acc.Balance, big.NewInt(1))
	assert.NoError(acc.SyntacticVerify())
	assert.ErrorContains(tx.Test(acc.Bytes()),
		`TxTester.Test error: test operation for path "/b" failed, expected "{false [1000000000]}", got "{false [1000000001]}"`)

	// TODO test LedgerObject

	// ModelObject
	tx = &TxTester{
		ObjectType: ModelObject,
		ObjectID:   CBORModelID.String(),
		Tests: TestOps{
			{Path: "/n", Value: util.MustMarshalCBOR("NameService")},
			{Path: "/th", Value: util.MustMarshalCBOR(uint64(1))},
			{Path: "/kp/0", Value: util.MustMarshalCBOR(util.Signer1.Address())},
			{Path: "/ap", Value: util.MustMarshalCBOR(nil)},
		},
	}
	assert.NoError(tx.SyntacticVerify())
	assert.False(tx.maybeTestData())

	data, err = json.Marshal(tx)
	assert.NoError(err)
	// fmt.Println(string(data))
	assert.Equal(`{"objectType":"Model","objectID":"1111111111111111111Ax1asG","tests":[{"path":"/n","value":"0x6b4e616d65536572766963655f6906be"},{"path":"/th","value":"0x017785459a"},{"path":"/kp/0","value":"0x548db97c7cece249c2b98bdc0226cc4c2a57bf52fc442832b9"},{"path":"/ap","value":"0xf65d4e5f13"}]}`, string(data))

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
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      []byte(sch),
	}
	assert.NoError(mi.SyntacticVerify())
	assert.ErrorContains(tx.Test(mi.Bytes()),
		`TxTester.Test error: test operation for path "/th" failed, expected "1", got "0"`)

	mi = &ModelInfo{
		Name:      "NameService",
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      []byte(sch),
	}
	assert.NoError(mi.SyntacticVerify())
	assert.NoError(tx.Test(mi.Bytes()))

	// DataObject
	tx = &TxTester{
		ObjectType: DataObject,
		ObjectID:   util.DataID{1, 2, 3}.String(),
		Tests: TestOps{
			{Path: "/v", Value: util.MustMarshalCBOR(uint64(1))},
			{Path: "/th", Value: util.MustMarshalCBOR(uint64(1))},
			{Path: "/kp/0", Value: util.MustMarshalCBOR(util.Signer1.Address())},
			{Path: "/ap", Value: util.MustMarshalCBOR(util.Signer2.Address())},
			{Path: "/d", Value: util.MustMarshalCBOR([]byte(`42`))},
		},
	}
	assert.NoError(tx.SyntacticVerify())
	assert.False(tx.maybeTestData())

	approver := util.Signer2.Address()
	di := &DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Approver:  &approver,
		Data:      []byte(`42`),
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(tx.Test(di.Bytes()))

	tx.Tests = append(tx.Tests[:len(tx.Tests)-1],
		TestOp{Path: "/d/name", Value: util.MustMarshalCBOR("John")},
		TestOp{Path: "/d/age", Value: util.MustMarshalCBOR(42)},
	)
	assert.NoError(tx.SyntacticVerify())
	assert.True(tx.maybeTestData())

	data, err = json.Marshal(tx)
	assert.NoError(err)
	// fmt.Println(string(data))
	assert.Equal(`{"objectType":"Data","objectID":"SkB7qHwfMsyF2PgrjhMvtFxJKhuR5ZfVoW9VATWRV4P9jV7J","tests":[{"path":"/v","value":"0x017785459a"},{"path":"/th","value":"0x017785459a"},{"path":"/kp/0","value":"0x548db97c7cece249c2b98bdc0226cc4c2a57bf52fc442832b9"},{"path":"/ap","value":"0x5444171c37ff5d7b7bb8dcad5c81f16284a229e641acaf799f"},{"path":"/d/name","value":"0x644a6f686e52bb61ab"},{"path":"/d/age","value":"0x182a20395c53"}]}`, string(data))

	type person struct {
		Name string `cbor:"name" json:"name"`
		Age  uint   `cbor:"age" json:"age"`
	}

	v := &person{Name: "John", Age: 42}
	di = &DataInfo{
		ModelID:   CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Approver:  &approver,
		Data:      util.MustMarshalCBOR(v),
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(tx.Test(di.Bytes()))

	di = &DataInfo{
		ModelID:   JSONModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Approver:  &approver,
		Data:      MustMarshalJSON(v),
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(tx.Test(di.Bytes()))
}
