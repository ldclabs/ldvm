// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package constants

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

var (
	// pjjsfTNAgQnP7zdpKfRcmicXGbk87xXznJmJZtqDAyRaNEhEL
	LDVMID = ids.ID{'l', 'd', 'v', 'm'}
	// 111111111111111111116DBWJs
	BlackholeAddr = ids.ShortEmpty
	// QLbz7JHiBTspS962RLKV8GndWFwdYhk6V
	GenesisAddr = ids.ShortID{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	}
	// LM111111111111111111116DBWJs
	RawModelID = ld.ModelID(ids.ShortEmpty)
	// LM1111111111111111111Ax1asG
	JsonModelID = ld.ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}
)
