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
	LDCAccount = ids.ShortEmpty
	// QLbz7JHiBTspS962RLKV8GndWFwdYhk6V
	// 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF
	GenesisAccount = ids.ShortID{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	}
	// LM111111111111111111116DBWJs
	RawModelID = util.ModelID(ids.ShortEmpty)
	// LM1111111111111111111Ax1asG
	JsonModelID = util.ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}
)
