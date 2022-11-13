// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"strings"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

const (
	AddressObject ObjectType = iota
	LedgerObject
	ModelObject
	DataObject
	// we will support testing trust data from outside
)

type ObjectType uint16

func (t ObjectType) String() string {
	switch t {
	case AddressObject:
		return "Address"
	case LedgerObject:
		return "Address"
	case ModelObject:
		return "Model"
	case DataObject:
		return "Data"
	default:
		return fmt.Sprintf("UnknownObjectType(%d)", t) // TODO: support for external data sources
	}
}

func (t ObjectType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.String() + "\""), nil
}

// TxTester
type TxTester struct {
	ObjectType ObjectType `cbor:"ot" json:"objectType"`
	ObjectID   string     `cbor:"oid" json:"objectID"`
	Tests      TestOps    `cbor:"ts" json:"tests"`

	// external assignment fields
	ID32 ids.ID32 `cbor:"-" json:"-"`
	ID20 ids.ID20 `cbor:"-" json:"-"`
	raw  []byte   `cbor:"-" json:"-"`
}

type TestOp struct {
	_ struct{} `cbor:",toarray"`

	Path  string           `json:"path"`
	Value encoding.RawData `json:"value"`
}

type TestOps []TestOp

func (ts TestOps) SyntacticVerify() error {
	errp := erring.ErrPrefix("ld.TestOps.SyntacticVerify: ")
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
	errp := erring.ErrPrefix("ld.TxTester.SyntacticVerify: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case len(t.Tests) == 0:
		return errp.Errorf("empty tests")
	}

	switch t.ObjectType {
	case AddressObject, LedgerObject:
		id, err := ids.AddressFromStr(t.ObjectID)
		if err != nil {
			return errp.ErrorIf(err)
		}
		t.ID20 = ids.ID20(id)

	case ModelObject:
		id, err := ids.ModelIDFromStr(t.ObjectID)
		if err != nil {
			return errp.ErrorIf(err)
		}
		t.ID20 = ids.ID20(id)

	case DataObject:
		id, err := ids.DataIDFromStr(t.ObjectID)
		if err != nil {
			return errp.ErrorIf(err)
		}
		t.ID32 = ids.ID32(id)

	default:
		return errp.Errorf("invalid objectType %s", t.ObjectType.String())
	}

	var err error
	if err = t.Tests.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *TxTester) maybeTestData() bool {
	if t.ObjectType == DataObject {
		for _, te := range t.Tests {
			if strings.HasPrefix(te.Path, "/pl/") {
				return true
			}
		}
	}
	return false
}

var rawRawModelID = string(encoding.MustMarshalCBOR(RawModelID))
var rawJSONModelID = string(encoding.MustMarshalCBOR(JSONModelID))

func (t *TxTester) Test(doc []byte) error {
	var err error

	errp := erring.ErrPrefix("ld.TxTester.Test: ")
	node := cborpatch.NewNode(doc)
	opts := cborpatch.NewOptions()

	if t.maybeTestData() {
		if rawModelID, _ := node.GetValue("/m", opts); rawModelID != nil {
			if rawData, _ := node.GetValue("/pl", opts); rawData != nil {
				var data []byte
				err = encoding.UnmarshalCBOR(rawData, &data)
				if err == nil {
					// try unwrap cbor data for testing
					switch string(rawModelID) {
					case rawRawModelID:
						// nothing to do
					case rawJSONModelID:
						data, err = cborpatch.FromJSON(data, nil)
						if err == nil {
							err = node.Patch(cborpatch.Patch{{Op: "replace", Path: "/pl", Value: data}}, opts)
						}
					default:
						err = node.Patch(cborpatch.Patch{{Op: "replace", Path: "/pl", Value: data}}, opts)
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
	return erring.ErrPrefix("ld.TxTester.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, t))
}

func (t *TxTester) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.TxTester.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(t))
}
