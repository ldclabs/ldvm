// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"encoding/json"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/units"

	"github.com/ldclabs/ldvm/ld"
)

type Config struct {
	FeeRecipient   ld.EthID       `json:"feeRecipient"`
	EventCacheSize int            `json:"eventCacheSize"`
	Logger         logging.Config `json:"logger"`
}

func New(data []byte) (*Config, error) {
	cfg := &Config{EventCacheSize: 100, Logger: logging.DefaultConfig}
	cfg.Logger.FileSize = 64 * units.MiB
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
