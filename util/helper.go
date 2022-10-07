// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"errors"
	"sort"

	"golang.org/x/crypto/sha3"
)

// Sum256 returns the SHA3-256 digest of the data.
func Sum256(msg []byte) []byte {
	d := sha3.Sum256(msg)
	return d[:]
}

func NodeIDToStakeAddress(nodeIDs ...Address) Addresses {
	rt := make(Addresses, len(nodeIDs))
	for i, id := range nodeIDs {
		rt[i] = id
		rt[i][0] = '#'
	}
	return rt
}

func HashFromData(data []byte) Hash {
	return Hash(sha3.Sum256(data))
}

type Addresses []Address

func (as Addresses) Has(id Address) bool {
	for _, v := range as {
		if v == id {
			return true
		}
	}
	return false
}

func (as Addresses) CheckEmptyID() error {
	for _, v := range as {
		if v == AddressEmpty {
			return errors.New("util.Addresses.CheckEmptyID: empty address exists")
		}
	}
	return nil
}

func (as Addresses) CheckDuplicate() error {
	set := make(map[string]struct{}, len(as))
	for _, v := range as {
		s := string(v[:])
		if _, ok := set[s]; ok {
			return errors.New("util.Addresses.CheckDuplicate: duplicate address " + v.String())
		}
		set[s] = struct{}{}
	}
	return nil
}

type DataIDs []DataID

func (s DataIDs) Has(id DataID) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}

func (s DataIDs) CheckEmptyID() error {
	for _, v := range s {
		if v == DataIDEmpty {
			return errors.New("util.DataIDs.CheckEmptyID: empty dataID exists")
		}
	}
	return nil
}

func (s DataIDs) CheckDuplicate() error {
	set := make(map[DataID]struct{}, len(s))
	for _, v := range s {
		if _, ok := set[v]; ok {
			return errors.New("util.Addresses.CheckDuplicate: duplicate dataID " + v.String())
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

func AddressToStakeSymbol(ids ...Address) []StakeSymbol {
	rt := make([]StakeSymbol, 0, len(ids))
	for _, id := range ids {
		rt = append(rt, id.ToStakeSymbol())
	}
	return rt
}
