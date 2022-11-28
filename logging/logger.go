// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package logging

import (
	avalogging "github.com/ava-labs/avalanchego/utils/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/util/value"
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

func MapToFields(m value.Map) []zapcore.Field {
	res := make([]zapcore.Field, 0, len(m))
	for _, k := range m.Keys() {
		res = append(res, zap.Any(k, m[k].ToAny()))
	}
	return res
}
