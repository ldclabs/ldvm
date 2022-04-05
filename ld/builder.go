// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"sync"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

func Recover(errfmt string, fn func()) (err error) {
	defer func() {
		ei := recover()
		switch ei.(type) {
		case nil:
			return
		default:
			err = fmt.Errorf("%s error: %v", errfmt, ei)
		}
	}()
	fn()
	return
}

type LDBuilder struct {
	mu         sync.Mutex
	name       string
	buf        *bytes.Buffer
	schemaType schema.Type
	prototype  schema.TypedPrototype
	builder    datamodel.NodeBuilder
}

func NewLDBuilder(name string, sch []byte, ptrType interface{}) (*LDBuilder, error) {
	b := &LDBuilder{name: name, buf: &bytes.Buffer{}}
	err := Recover("build "+name, func() {
		ts, err := ipld.LoadSchemaBytes(sch)
		if err != nil {
			panic(err)
		}
		b.schemaType = ts.TypeByName(name)
		b.prototype = bindnode.Prototype(ptrType, b.schemaType)
		b.builder = b.prototype.Representation().NewBuilder()
	})
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (l *LDBuilder) Name() string {
	return l.name
}

func (l *LDBuilder) Type() schema.Type {
	return l.schemaType
}

func (l *LDBuilder) Marshal(bind interface{}) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	defer l.buf.Reset()

	err := Recover("LDBuilder marshal "+l.name, func() {
		node := bindnode.Wrap(bind, l.schemaType)
		if err := dagcbor.Encode(node.Representation(), l.buf); err != nil {
			panic(err)
		}
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

	err := Recover("LDBuilder marshal json "+l.name, func() {
		node := bindnode.Wrap(bind, l.schemaType)
		if err := dagjson.Encode(node, l.buf); err != nil {
			panic(err)
		}
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

	err = Recover("LDBuilder marshal "+l.name, func() {
		builder := l.prototype.Representation().NewBuilder()
		if er := dagcbor.Decode(builder, bytes.NewReader(data)); er != nil {
			panic(err)
		}
		node := builder.Build()
		bind = bindnode.Unwrap(node)
	})
	if err == nil && bind == nil {
		err = fmt.Errorf("LDBuilder marshal %s error: Unwrap return nil", l.name)
	}
	return
}
