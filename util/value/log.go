// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package value

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Log repsents a log object base on Value.
type Log struct {
	Value
	Err error
}

func (l *Log) Valid() bool {
	return l != nil && l.Is(Vmap)
}

func DefaultLogHandler(log *Log) {
	var err error
	log.Set("time", Time(time.Now()))

	switch {
	case log.Err != nil:
		log.Set("error", String(log.Err.Error()))
		err = json.NewEncoder(os.Stderr).Encode(log.ToMap())
	default:
		err = json.NewEncoder(os.Stdout).Encode(log.ToMap())
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "HandleLog: %v, %s", err, log.GoString())
	}
}
