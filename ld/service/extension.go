// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/erring"
	"github.com/ldclabs/ldvm/util/validating"
)

type Extension struct {
	Title      string                 `cbor:"t" json:"title"` // extension title
	Properties map[string]interface{} `cbor:"ps" json:"properties"`
	DataID     *ids.DataID            `cbor:"id,omitempty" json:"did,omitempty"` // optional data id
	ModelID    *ids.ModelID           `cbor:"m,omitempty" json:"mid,omitempty"`  // optional model id
}

type Extensions []*Extension

// SyntacticVerify verifies that Extensions is well-formed.
func (es Extensions) SyntacticVerify() error {
	errp := erring.ErrPrefix("service.Extensions.SyntacticVerify: ")
	if es == nil {
		return errp.Errorf("nil pointer")
	}

	set := make(map[string]struct{}, len(es))
	for i, ex := range es {
		switch {
		case ex == nil:
			return errp.Errorf("nil pointer at %d", i)

		case !validating.ValidName(ex.Title):
			return errp.Errorf("invalid title %q at %d", ex.Title, i)

		case ex.Properties == nil:
			return errp.Errorf("nil properties at %d", i)

		case ex.DataID != nil && ex.ModelID == nil:
			return errp.Errorf("nil model id at %d", i)

		case ex.DataID == nil && ex.ModelID != nil:
			return errp.Errorf("no data id at %d, model id be nil", i)

		case ex.DataID != nil && *ex.DataID == ids.EmptyDataID:
			return errp.Errorf("invalid data id at %d", i)

		case ex.ModelID != nil && *ex.ModelID == ids.EmptyModelID:
			return errp.Errorf("invalid model id at %d", i)
		}

		key := ex.Title
		if ex.ModelID != nil {
			key = string((*ex.ModelID)[:]) + key
		}

		if _, ok := set[key]; ok {
			return errp.Errorf("%q exists in extensions at %d", ex.Title, i)
		}
		set[key] = struct{}{}
	}

	return nil
}
