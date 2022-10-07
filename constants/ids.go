// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package constants

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
)

var (
	// pjjsfTNAgQnP7zdpKfRcmicXGbk87xXznJmJZtqDAyRaNEhEL
	LDVMID = ids.ID{'l', 'd', 'v', 'm'}
	// 0x0000000000000000000000000000000000000000
	LDCAccount = util.AddressEmpty
	// 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff
	GenesisAccount = util.Address{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	}
	NativeToken = util.NativeToken
)
