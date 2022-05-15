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
	// 111111111111111111116DBWJs
	// 0x0000000000000000000000000000000000000000
	LDCAccount = util.EthIDEmpty
	// QLbz7JHiBTspS962RLKV8GndWFwdYhk6V
	// 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF
	GenesisAccount = util.EthID{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	}
	NativeToken = util.NativeToken
	// LM111111111111111111116DBWJs
	RawModelID = util.ModelIDEmpty
	// LM1111111111111111111Ax1asG
	CBORModelID = util.ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}
	// LM1111111111111111111L17Xp3
	JSONModelID = util.ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 2,
	}
)
