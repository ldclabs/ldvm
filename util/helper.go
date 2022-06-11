// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"golang.org/x/crypto/sha3"
)

func JSONMarshalData(data []byte) json.RawMessage {
	switch {
	case len(data) == 0 || json.Valid(data):
		return data
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
	if last := len(data) - 1; last > 2 && data[0] == '"' && data[last] == '"' {
		b := data[1:last]
		if d, err := formatting.Decode(formatting.Hex, string(b)); err == nil {
			return d
		}
	}
	return data
}

func NodeIDToStakeAddress(nodeIDs ...EthID) []EthID {
	rt := make([]EthID, len(nodeIDs))
	for i, id := range nodeIDs {
		rt[i] = id
		rt[i][0] = '#'
	}
	return rt
}

func IDFromData(data []byte) ids.ID {
	return ids.ID(sha3.Sum256(data))
}

type EthIDs []EthID

func (s EthIDs) Has(id EthID) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}

func (s EthIDs) CheckEmptyID() error {
	for _, v := range s {
		if v == EthIDEmpty {
			return fmt.Errorf("empty address exists")
		}
	}
	return nil
}

func (s EthIDs) CheckDuplicate() error {
	set := make(map[EthID]struct{}, len(s))
	for _, v := range s {
		if _, ok := set[v]; ok {
			return fmt.Errorf("duplicate address %s", v)
		}
		set[v] = struct{}{}
	}
	return nil
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
