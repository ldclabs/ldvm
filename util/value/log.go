// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package value

// Log repsents a log object base on Value.
type Log struct{ Value }

func (l *Log) Valid() bool {
	return l != nil && l.Is(Vmap)
}
