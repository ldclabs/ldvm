// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"sync"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	cborpatch "github.com/ldclabs/cbor-patch"

	"github.com/ldclabs/ldvm/util"
)

type IPLDModel struct {
	mu         sync.Mutex
	name       string
	schema     string
	buf        *bytes.Buffer
	schemaType schema.Type
	prototype  schema.TypedPrototype
	// builder    datamodel.NodeBuilder
}

func NewIPLDModel(name string, sc string) (*IPLDModel, error) {
	b := &IPLDModel{name: name, schema: sc, buf: new(bytes.Buffer)}

	errp := util.ErrPrefix(fmt.Sprintf("NewIPLDModel(%q) error: ", name))
	err := Recover(errp, func() error {
		ts, err := ipld.LoadSchemaBytes([]byte(sc))
		if err != nil {
			return err
		}

		b.schemaType = ts.TypeByName(name)
		switch typ := b.schemaType.(type) {
		case *schema.TypeMap, *schema.TypeList, *schema.TypeStruct:
		case nil:
			return fmt.Errorf("type not found")
		default:
			return fmt.Errorf("should be a map, list or struct, but got %s", typ.TypeKind())
		}

		b.prototype = bindnode.Prototype(nil, b.schemaType)
		// b.builder = b.prototype.Representation().NewBuilder()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return b, nil
}

func (l *IPLDModel) Name() string {
	return l.name
}

func (l *IPLDModel) Schema() string {
	return l.schema
}

func (l *IPLDModel) Type() schema.Type {
	return l.schemaType
}

func (l *IPLDModel) Decode(doc []byte) (node datamodel.Node, err error) {
	errp := util.ErrPrefix(fmt.Sprintf("IPLDModel(%q).Decode error: ", l.name))

	l.mu.Lock()
	defer l.mu.Unlock()

	node, err = l.decode(doc)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	return node, nil
}

func (l *IPLDModel) ApplyPatch(doc, operations []byte) ([]byte, error) {
	errp := util.ErrPrefix(fmt.Sprintf("IPLDModel(%q).ApplyPatch error: ", l.name))

	p, err := cborpatch.NewPatch(operations)
	if err != nil {
		return nil, errp.Errorf("invalid CBOR patch, %v", err)
	}

	if doc, err = p.Apply(doc); err != nil {
		return nil, errp.ErrorIf(err)
	}

	if err = l.valid(doc); err != nil {
		return nil, errp.ErrorIf(err)
	}
	return doc, nil
}

func (l *IPLDModel) Valid(data []byte) error {
	errp := util.ErrPrefix(fmt.Sprintf("IPLDModel(%q).Valid error: ", l.name))
	if err := util.ValidCBOR(data); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(l.valid(data))
}

func (l *IPLDModel) valid(data []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	node, err := l.decode(data)
	if err != nil {
		return err
	}

	defer l.buf.Reset()
	if err = dagcbor.Encode(node, l.buf); err != nil {
		return err
	}
	if d := l.buf.Bytes(); !bytes.Equal(data, d) {
		err = fmt.Errorf("data not equal, length expected %v, got %v",
			len(data), len(d))
	}
	return err
}

func (l *IPLDModel) decode(doc []byte) (node datamodel.Node, err error) {
	// defer l.builder.Reset() TODO: not supported yet
	errp := util.ErrPrefix("decode error: ")
	err = Recover(errp, func() error {
		builder := l.prototype.Representation().NewBuilder()
		if er := dagcbor.Decode(builder, bytes.NewReader(doc)); er != nil {
			return er
		}
		node = builder.Build()
		if tn, ok := node.(schema.TypedNode); ok {
			node = tn.Representation()
		}
		return nil
	})

	if err == nil && node == nil {
		err = errp.Errorf("%d bytes return nil", len(doc))
	}
	return
}
