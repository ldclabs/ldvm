// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

type LDBuilder struct {
	mu         sync.Mutex
	name       string
	sch        []byte
	buf        *bytes.Buffer
	schemaType schema.Type
	prototype  schema.TypedPrototype
	builder    datamodel.NodeBuilder
}

func NewLDBuilder(name string, sch []byte, ptrType interface{}) (*LDBuilder, error) {
	b := &LDBuilder{name: name, sch: sch, buf: &bytes.Buffer{}}
	err := Recover("build "+name, func() error {
		ts, err := ipld.LoadSchemaBytes(sch)
		if err != nil {
			return err
		}
		b.schemaType = ts.TypeByName(name)
		b.prototype = bindnode.Prototype(ptrType, b.schemaType)
		b.builder = b.prototype.Representation().NewBuilder()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (l *LDBuilder) Name() string {
	return l.name
}

func (l *LDBuilder) Schema() []byte {
	return l.sch
}

func (l *LDBuilder) Type() schema.Type {
	return l.schemaType
}

func (l *LDBuilder) Marshal(bind interface{}) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	defer l.buf.Reset()

	err := Recover("LDBuilder marshal "+l.name, func() error {
		node := bindnode.Wrap(bind, l.schemaType)
		return dagcbor.Encode(node.Representation(), l.buf)
	})

	if err != nil {
		return nil, err
	}

	data := make([]byte, l.buf.Len())
	copy(data, l.buf.Bytes())
	return data, nil
}

func (l *LDBuilder) ToJSON(bind interface{}) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	defer l.buf.Reset()

	err := Recover("LDBuilder marshal json "+l.name, func() error {
		node := bindnode.Wrap(bind, l.schemaType)
		return dagjson.Encode(node, l.buf)
	})

	if err != nil {
		return nil, err
	}

	data := make([]byte, l.buf.Len())
	copy(data, l.buf.Bytes())
	return data, nil
}

func (l *LDBuilder) Unmarshal(data []byte) (bind interface{}, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// defer l.builder.Reset() TODO: not supported yet

	err = Recover("LDBuilder marshal "+l.name, func() error {
		builder := l.prototype.Representation().NewBuilder()
		if er := dagcbor.Decode(builder, bytes.NewReader(data)); er != nil {
			return er
		}
		node := builder.Build()
		bind = bindnode.Unwrap(node)
		return nil
	})
	if err == nil && bind == nil {
		err = fmt.Errorf("LDBuilder marshal %s error: Unwrap return nil", l.name)
	}
	return
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

func WriteBytesList(w io.Writer, b []byte, bb ...[]byte) error {
	return Recover("WriteBytesList", func() error {
		nb := basicnode.Prototype.List.NewBuilder()
		la, er := nb.BeginList(int64(len(bb) + 1))
		if er != nil {
			return er
		}
		la.AssembleValue().AssignBytes(b)
		for _, v := range bb {
			la.AssembleValue().AssignBytes(v)
		}
		if er = la.Finish(); er != nil {
			return er
		}
		return dagcbor.Encode(nb.Build(), w)
	})
}

func ReadBytesList(r io.Reader) ([][]byte, error) {
	var err error
	nb := basicnode.Prototype.List.NewBuilder()
	if err = dagcbor.Decode(nb, r); err != nil {
		return nil, err
	}

	node := nb.Build()
	ln := node.Length()
	list := make([][]byte, ln)

	var n ipld.Node
	var b []byte
	for i := int64(0); i < ln; i++ {
		n, err = node.LookupByIndex(i)
		if err != nil {
			return nil, err
		}
		b, err = n.AsBytes()
		if err != nil {
			return nil, err
		}
		list[i] = b
	}
	return list, nil
}
