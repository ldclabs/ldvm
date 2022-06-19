// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/ldclabs/ldvm/util"
)

type IPLDModel struct {
	mu         sync.Mutex
	name       string
	sch        []byte
	buf        *bytes.Buffer
	schemaType schema.Type
	prototype  schema.TypedPrototype
	// builder    datamodel.NodeBuilder
}

func NewIPLDModel(name string, sch []byte) (*IPLDModel, error) {
	b := &IPLDModel{name: name, sch: sch, buf: new(bytes.Buffer)}

	errp := util.ErrPrefix(fmt.Sprintf("NewIPLDModel(%s) error: ", strconv.Quote(name)))
	err := Recover(errp, func() error {
		ts, err := ipld.LoadSchemaBytes(sch)
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

func (l *IPLDModel) Schema() []byte {
	return l.sch
}

func (l *IPLDModel) Type() schema.Type {
	return l.schemaType
}

func (l *IPLDModel) decode(data []byte) (node datamodel.Node, err error) {
	// defer l.builder.Reset() TODO: not supported yet

	errp := util.ErrPrefix(fmt.Sprintf("IPLDModel(%s).decode error: ", strconv.Quote(l.name)))
	err = Recover(errp, func() error {
		builder := l.prototype.Representation().NewBuilder()
		if er := dagcbor.Decode(builder, bytes.NewReader(data)); er != nil {
			return er
		}
		node = builder.Build()
		if tn, ok := node.(schema.TypedNode); ok {
			node = tn.Representation()
		}
		return nil
	})
	if err == nil && node == nil {
		err = errp.Errorf("%d bytes return nil", len(data))
	}
	return
}

func (l *IPLDModel) ApplyPatch(original, patch []byte) ([]byte, error) {
	return nil, fmt.Errorf("IPLDModel.ApplyPatch TODO")
}

func (l *IPLDModel) Valid(data []byte) error {
	errp := util.ErrPrefix(fmt.Sprintf("IPLDModel(%s).Valid error: ", strconv.Quote(l.name)))
	if err := util.ValidCBOR(data); err != nil {
		return errp.ErrorIf(err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	node, err := l.decode(data)
	if err != nil {
		return err
	}

	defer l.buf.Reset()
	if err = dagcbor.Encode(node, l.buf); err != nil {
		return errp.ErrorIf(err)
	}
	d := l.buf.Bytes()
	if !bytes.Equal(data, d) {
		err = errp.Errorf("data not equal, bytes length expected %v, got %v",
			len(data), len(d))
	}
	return err
}
