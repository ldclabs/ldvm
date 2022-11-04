// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"golang.org/x/crypto/sha3"
)

// Sum256 returns the SHA3-256 digest of the data.
func Sum256(msg []byte) []byte {
	d := sha3.Sum256(msg)
	return d[:]
}

func HashFromData(data []byte) Hash {
	return Hash(sha3.Sum256(data))
}

func AddressToStakeSymbol(ids ...Address) IDList[StakeSymbol] {
	rt := make([]StakeSymbol, 0, len(ids))
	for _, id := range ids {
		rt = append(rt, id.ToStakeSymbol())
	}
	return rt
}
