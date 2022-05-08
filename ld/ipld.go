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
	err := Recover("NewIPLDModel "+name, func() error {
		ts, err := ipld.LoadSchemaBytes(sch)
		if err != nil {
			return err
		}
		b.schemaType = ts.TypeByName(name)
		switch typ := b.schemaType.(type) {
		case *schema.TypeMap, *schema.TypeList, *schema.TypeStruct:
		default:
			return fmt.Errorf("model should be a map, list or struct, but got %v", typ)
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

	err = Recover("IPLDModel "+l.name, func() error {
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
		err = fmt.Errorf("IPLDModel %s decode node error: %d bytes return nil", l.name, len(data))
	}
	return
}

func (l *IPLDModel) ApplyPatch(original, patch []byte) ([]byte, error) {
	return nil, fmt.Errorf("*IPLDModel.ApplyPatch TODO")
}

func (l *IPLDModel) Valid(data []byte) error {
	if err := DecMode.Valid(data); err != nil {
		return err
	}

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
	d := l.buf.Bytes()
	if !bytes.Equal(data, d) {
		err = fmt.Errorf("Model(%s) valid data failed, data length expected %v, got %v",
			strconv.Quote(l.name), len(data), len(d))
	}
	return err
}

func NewSchemaType(name string, sch []byte) (schema.Type, error) {
	var st schema.Type
	err := Recover("build "+name, func() error {
		ts, er := ipld.LoadSchemaBytes(sch)
		if er != nil {
			return er
		}
		st = ts.TypeByName(name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return st, nil
}
