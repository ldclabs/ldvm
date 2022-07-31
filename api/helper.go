// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

// -32000 to -32099	Server error	Reserved for implementation-defined server-errors.
const (
	CodeServerError = -32000 - iota
	CodeAccountError
)

func formatEthBalance(amount *big.Int) string {
	return "0x" + ld.ToEthBalance(amount).Text(16)
}

func formatUint64(u uint64) string {
	return "0x" + strconv.FormatUint(u, 16)
}

func formatBytes(data []byte) string {
	return "0x" + hex.EncodeToString(data)
}

func decodeBytes(s string) ([]byte, error) {
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}

	return hex.DecodeString(s)
}

func decodeAddress(s string) (id util.EthID, err error) {
	data, e := decodeBytes(s)
	if e != nil {
		err = e
		return
	}

	if len(data) != 20 {
		err = fmt.Errorf("invalid address, %s", s)
		return
	}
	copy(id[:], data)
	return
}

func decodeHash(s string) (id ids.ID, err error) {
	data, e := decodeBytes(s)
	if e != nil {
		err = e
		return
	}

	if len(data) != 32 {
		err = fmt.Errorf("invalid hash, %s", s)
		return
	}
	copy(id[:], data)
	return
}

func decodeHashByRaw(data json.RawMessage) (id ids.ID, err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	return decodeHash(s)
}
