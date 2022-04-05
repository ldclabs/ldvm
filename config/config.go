// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"errors"
)

type Config struct {
}

func (c *Config) SetDefaults() {
}

func New(data []byte) (*Config, error) {
	return nil, errors.New("TODO")
}
