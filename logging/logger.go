// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package logging

import (
	"runtime"

	avalogging "github.com/ava-labs/avalanchego/utils/logging"

	"github.com/ldclabs/ldvm/config"
)

var Log avalogging.Logger = &avalogging.NoLog{}
var cfg avalogging.Config

func init() {
	cfg = config.DefaultLoggingConfig
	logFactory := avalogging.NewFactory(cfg)
	Log, _ = logFactory.Make("ldvm")
}

func Debug(fn func() string) {
	if cfg.LogLevel <= avalogging.Debug {
		Log.Debug(fn())
	}
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
