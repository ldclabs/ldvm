// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"unicode/utf8"

	"github.com/ava-labs/avalanchego/utils/formatting"
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

	if err = fn(); err != nil {
		return fmt.Errorf("%s error: %v", errfmt, err)
	}
	return nil
}

func JSONMarshalData(data []byte) json.RawMessage {
	switch {
	case len(data) == 0 || json.Valid(data):
		return data
	case utf8.Valid(data):
		return []byte(strconv.Quote(string(data)))
	default:
		s, err := formatting.EncodeWithChecksum(formatting.CB58, data)
		if err != nil {
			return data
		}
		buf := make([]byte, len(s)+2)
		buf[0] = '"'
		copy(buf[1:], []byte(s))
		buf[len(buf)-1] = '"'
		return buf
	}
}

func JSONUnmarshalData(data json.RawMessage) []byte {
	if last := len(data) - 1; last > 10 && data[0] == '"' && data[last] == '"' {
		if d, err := formatting.Decode(formatting.CB58, string(data[1:last])); err == nil {
			return d
		}
	}
	return data
}
