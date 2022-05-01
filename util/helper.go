// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/json"
	"sort"
	"strconv"
	"unicode/utf8"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"golang.org/x/crypto/sha3"
)

var Null = []byte("null")

func JSONMarshalData(data []byte) json.RawMessage {
	switch {
	case len(data) == 0 || json.Valid(data):
		return data
	case utf8.Valid(data):
		return []byte(strconv.Quote(string(data)))
	default:
		s, err := formatting.EncodeWithChecksum(formatting.Hex, data)
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
		if d, err := formatting.Decode(formatting.Hex, string(data[1:last])); err == nil {
			return d
		}
	}
	return data
}

func NodeIDToStakeAddress(nodeIDs ...ids.ShortID) []ids.ShortID {
	rt := make([]ids.ShortID, len(nodeIDs))
	for i, id := range nodeIDs {
		rt[i] = id
		rt[i][0] = '$'
	}
	return rt
}

func IDFromBytes(data []byte) ids.ID {
	return ids.ID(sha3.Sum256(data))
}

type ShortIDs []ids.ShortID

func (s ShortIDs) Has(id ids.ShortID) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}

type Uint64Set map[uint64]struct{}

func (us Uint64Set) Has(u uint64) bool {
	_, ok := us[u]
	return ok
}

func (us Uint64Set) Add(uu ...uint64) {
	for _, u := range uu {
		us[u] = struct{}{}
	}
}

func (us Uint64Set) List() []uint64 {
	list := make([]uint64, 0, len(us))
	for u := range us {
		list = append(list, u)
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i] < list[j] })
	return list
}
