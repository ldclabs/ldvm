// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/genesis"
)

type Context struct {
	*snow.Context
	state   StateDB
	config  *config.Config
	genesis *genesis.Genesis
}

func NewContext(
	ctx *snow.Context,
	state StateDB,
	config *config.Config,
	genesis *genesis.Genesis,
) *Context {
	return &Context{ctx, state, config, genesis}
}

func (c *Context) StateDB() StateDB {
	return c.state
}

func (c *Context) Chain() *genesis.ChainConfig {
	return &c.genesis.Chain
}

func (c *Context) Config() *config.Config {
	return c.config
}
