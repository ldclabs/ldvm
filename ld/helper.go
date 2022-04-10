// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"unicode/utf8"
)

var Null = []byte("null")

func Recover(errfmt string, fn func() error) (err error) {
	defer func() {
		if re := recover(); re != nil {
			buf := make([]byte, 2048)
			buf = buf[:runtime.Stack(buf, false)]
			err = fmt.Errorf("%s panic: %v, stack: %s", errfmt, re, string(buf))
		}
	}()
	return fn()
}

func JsonMarshalData(data []byte) json.RawMessage {
	switch {
	case len(data) == 0 || json.Valid(data):
		return data
	case utf8.Valid(data):
		return []byte(strconv.Quote(string(data)))
	default:
		buf := make([]byte, base64.StdEncoding.EncodedLen(len(data))+2)
		buf[0] = '"'
		base64.StdEncoding.Encode(buf[1:], data)
		buf[len(buf)-1] = '"'
		return buf
	}
}
