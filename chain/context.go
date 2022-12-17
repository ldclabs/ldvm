// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ldclabs/cose/key"

	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ids"
)

type Context struct {
	*snow.Context
	bc      BlockChain
	config  *config.Config
	genesis *genesis.Genesis
	name    string
	builder ids.Address
	signer  key.Signer
}

func NewContext(
	name string,
	ctx *snow.Context,
	bc BlockChain,
	config *config.Config,
	genesis *genesis.Genesis,
) *Context {
	cc := &Context{ctx, bc, config, genesis, name, ids.Address(ctx.NodeID), nil}
	if config.Builder != nil {
		cc.builder = config.Builder.Address
		cc.signer = config.Builder.Signer
	}
	return cc
}

func (c *Context) Name() string {
	return c.name
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

func (c *Context) Builder() ids.Address {
	return c.builder
}

func (c *Context) BuilderSigner() key.Signer {
	return c.signer
}
