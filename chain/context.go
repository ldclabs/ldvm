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
	bc      BlockChain
	config  *config.Config
	genesis *genesis.Genesis
}

func NewContext(
	ctx *snow.Context,
	bc BlockChain,
	config *config.Config,
	genesis *genesis.Genesis,
) *Context {
	return &Context{ctx, bc, config, genesis}
}

func (c *Context) Chain() BlockChain {
	return c.bc
}

func (c *Context) Config() *config.Config {
	return c.config
}

func (c *Context) ChainConfig() *genesis.ChainConfig {
	return &c.genesis.Chain
}
