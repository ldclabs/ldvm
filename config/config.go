// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/units"
)

var (
	homeDir              = os.ExpandEnv("$HOME")
	DefaultLogDirectory  = fmt.Sprintf("%s/.%s/logs", homeDir, constants.AppName)
	DefaultLoggingConfig = logging.Config{
		RotatingWriterConfig: logging.RotatingWriterConfig{
			MaxSize:   64 * units.MiB,
			MaxFiles:  32,
			MaxAge:    7,
			Directory: DefaultLogDirectory,
			Compress:  false,
		},
		DisplayLevel: logging.Info,
		LogLevel:     logging.Debug,
	}
)

type Config struct {
	EventCacheSize int            `json:"eventCacheSize"`
	Logger         logging.Config `json:"logger"`
}

func New(data []byte) (*Config, error) {
	cfg := &Config{EventCacheSize: 100, Logger: DefaultLoggingConfig}
	if len(data) > 0 {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}
	if cfg.EventCacheSize <= 0 {
		cfg.EventCacheSize = 100
	}
	return cfg, nil
}
