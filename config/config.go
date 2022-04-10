// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"encoding/json"

	"github.com/ava-labs/avalanchego/utils/logging"

	"github.com/ldclabs/ldvm/ld"
)

type Config struct {
	FeeRecipient     ld.EthID       `json:"feeRecipient"`
	RecentEventsSize int            `json:"recentEventsSize"`
	Logger           logging.Config `json:"logger"`
}

func New(data []byte) (*Config, error) {
	cfg := &Config{RecentEventsSize: 100, Logger: logging.DefaultConfig}
	if len(data) > 0 {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}
	if cfg.RecentEventsSize <= 0 {
		cfg.RecentEventsSize = 100
	}
	return cfg, nil
}
