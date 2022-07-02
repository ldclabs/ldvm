// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"strings"

	"github.com/ava-labs/avalanchego/ids"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

const (
	AddressObject ObjectType = iota
	ModelObject
	DataObject
)

type ObjectType uint16

func (t ObjectType) String() string {
	switch t {
	case AddressObject:
		return "Address"
	case ModelObject:
		return "Model"
	case DataObject:
		return "Data"
	default:
		return fmt.Sprintf("UnknownObjectType(%d)", t)
	}
}

func (t ObjectType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.String() + "\""), nil
}

// TxTester
type TxTester struct {
	ObjectType ObjectType  `cbor:"ot" json:"objectType"`
	OID        any         `cbor:"-" json:"objectId"` // external field
	ObjectID   ids.ShortID `cbor:"oid" json:"-"`
	Tests      TestOps     `cbor:"ts" json:"tests"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

type TestOp struct {
	_ struct{} `cbor:",toarray"`

	Path  string       `json:"path"`
	Value util.RawData `json:"value"`
}

type TestOps []TestOp

func (ts TestOps) SyntacticVerify() error {
	errp := util.ErrPrefix("TestOps.SyntacticVerify error: ")
	for _, t := range ts {
		switch {
		case t.Path == "":
			return errp.Errorf("invalid path")
		case len(t.Value) == 0:
			return errp.Errorf("invalid value")
		}
	}
	return nil
}

func (ts TestOps) ToPatch() cborpatch.Patch {
	p := make(cborpatch.Patch, len(ts))
	for i, t := range ts {
		p[i] = cborpatch.Operation{
			Op:    "test",
			Path:  t.Path,
			Value: cborpatch.RawMessage(t.Value),
		}
	}
	return p
}

// SyntacticVerify verifies that a *TxTester is well-formed.
func (t *TxTester) SyntacticVerify() error {
	errp := util.ErrPrefix("TxTester.SyntacticVerify error: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case t.ObjectType > DataObject:
		return errp.Errorf("invalid objectType %s", t.ObjectType.String())

	case len(t.Tests) == 0:
		return errp.Errorf("empty tests")
	}

	var err error
	if err = t.Tests.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch t.ObjectType {
	case AddressObject:
		t.OID = util.EthID(t.ObjectID)
	case ModelObject:
		t.OID = util.ModelID(t.ObjectID)
	case DataObject:
		t.OID = util.DataID(t.ObjectID)
	}

	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *TxTester) maybeTestData() bool {
	if t.ObjectType == DataObject {
		for _, te := range t.Tests {
			if strings.HasPrefix(te.Path, "/d/") {
				return true
			}
		}
	}
	return false
}

var rawRawModelID = string(util.MustMarshalCBOR(constants.RawModelID))
var rawJSONModelID = string(util.MustMarshalCBOR(constants.JSONModelID))

func (t *TxTester) Test(doc []byte) error {
	var err error

	errp := util.ErrPrefix("TxTester.Test error: ")
	node := cborpatch.NewNode(doc)
	opts := cborpatch.NewOptions()

	if t.maybeTestData() {
		if rawModelID, _ := node.GetValue("/m", opts); rawModelID != nil {
			if rawData, _ := node.GetValue("/d", opts); rawData != nil {
				var data []byte
				err = util.UnmarshalCBOR(rawData, &data)
				if err == nil {
					// try unwrap cbor data for testing
					switch string(rawModelID) {
					case rawRawModelID:
						// nothing to do
					case rawJSONModelID:
						fmt.Println(string(data))
						data, err = cborpatch.FromJSON(data, nil)
						if err == nil {
							fmt.Println(cborpatch.MustToJSON(data))
							err = node.Patch(cborpatch.Patch{{Op: "replace", Path: "/d", Value: data}}, opts)
						}
					default:
						fmt.Println(cborpatch.MustToJSON(data))
						err = node.Patch(cborpatch.Patch{{Op: "replace", Path: "/d", Value: data}}, opts)
					}
				}

				if err != nil {
					return errp.ErrorIf(err)
				}
			}
		}
	}

	return errp.ErrorIf(node.Patch(t.Tests.ToPatch(), opts))
}

func (t *TxTester) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxTester) Unmarshal(data []byte) error {
	return util.ErrPrefix("TxTester.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *TxTester) Marshal() ([]byte, error) {
	return util.ErrPrefix("TxTester.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}
