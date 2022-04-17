// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package logging

import (
	"runtime"

	avalogging "github.com/ava-labs/avalanchego/utils/logging"
)

var Log avalogging.Logger = &avalogging.NoLog{}

func init() {
	logFactory := avalogging.NewFactory(avalogging.DefaultConfig)
	Log, _ = logFactory.Make("ldvm")
}

func SetLogger(l avalogging.Logger) {
	Log.Stop()
	Log = l
}

func LogStack(format string, args ...interface{}) {
	buf := make([]byte, 2048)
	buf = buf[:runtime.Stack(buf, false)]
	format += "\nstack: %s"
	args = append(args, string(buf))
	Log.Info(format, args...)
}
