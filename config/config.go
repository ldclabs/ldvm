// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	avaids "github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ldclabs/cose/key"
	"github.com/ldclabs/cose/key/ed25519"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/encoding"
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
	Logger      logging.Config `json:"logger"`
	RPCAddr     string         `json:"rpcAddr"`
	POSEndpoint string         `json:"posEndpoint"` // persistent data source endpoint
	Builder     *Builder       `json:"builder"`
}

type Builder struct {
	Node          avaids.NodeID `json:"nodeId"`
	Address       ids.Address   `json:"address"`
	PrivateSeed   string        `json:"privateSeed"` // optional, only for local test
	KesEndpoint   string        `json:"kesEndpoint"`
	KesKeyName    string        `json:"kesKeyName"`
	KesCipherText string        `json:"kesCipherText"`
	KesSignature  string        `json:"kesSignature"`

	Signer key.Signer `json:"-"`
}

func (b *Builder) Valid(ctx context.Context) error {
	if b.PrivateSeed != "" {
		privateSeed, err := encoding.DecodeString(b.PrivateSeed)
		if err != nil {
			return err
		}
		key, err := ed25519.KeyFromSeed(privateSeed)
		if err != nil {
			return err
		}
		b.Signer, err = key.Signer()
		if err != nil {
			return err
		}
	}

	// TODO: read private seed from KES
	return nil
}

func New(ctx context.Context, data []byte) (*Config, error) {
	cfg := &Config{
		Logger:      DefaultLoggingConfig,
		RPCAddr:     ":2357",
		POSEndpoint: "h2c://localhost:2357/pos",
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	if cfg.Builder != nil {
		if err := cfg.Builder.Valid(ctx); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
