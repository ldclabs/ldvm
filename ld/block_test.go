// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
)

func TestBlock(t *testing.T) {
	assert := assert.New(t)

	var blk *Block
	assert.ErrorContains(blk.SyntacticVerify(), "nil pointer")

	blk = &Block{}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid state AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX")

	blk = &Block{State: util.Hash{1, 2, 3}, Timestamp: uint64(time.Now().Unix() + 20)}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid timestamp")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 1}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid gasPrice")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100, GasRebateRate: 1001}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid gasRebateRate")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100, Miner: util.StakeSymbol{1, 2, 3}}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid miner address")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100}
	assert.ErrorContains(blk.SyntacticVerify(), "no txs")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100, Txs: make([]*Transaction, 1),
		Validators: []util.StakeSymbol{{1, 2, 3}}}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid validator address")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100, Txs: make([]*Transaction, 1)}
	assert.ErrorContains(blk.SyntacticVerify(), "Transaction.SyntacticVerify: nil pointer")

	tx := &Transaction{
		Tx: TxData{
			Type:      TypeTransfer,
			ChainID:   gChainID,
			Nonce:     1,
			GasTip:    0,
			GasFeeCap: 1000,
			From:      signer.Signer1.Key().Address(),
			To:        signer.Signer2.Key().Address().Ptr(),
			Amount:    big.NewInt(1_000_000),
		},
	}
	assert.NoError(tx.SignWith(signer.Signer1))
	assert.NoError(tx.SyntacticVerify())

	blk = &Block{
		State:         util.Hash{1, 2, 3},
		Gas:           tx.Gas(),
		GasPrice:      1000,
		GasRebateRate: 200,
		Txs:           Txs{tx},
	}

	assert.NoError(blk.SyntacticVerify())
	cbordata, err := blk.Marshal()
	assert.NoError(err)

	assert.NoError(blk.TxsMarshalJSON())
	jsondata, err := json.Marshal(blk)
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"parent":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX","height":0,"timestamp":0,"state":"AQIDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAoWLSv","gas":638,"gasPrice":1000,"gasRebateRate":200,"miner":"","validators":null,"pChainHeight":0,"id":"bZ2rHS4yOz7W6lbLwmLKGqE6w3wRPlNzt5u4kEbdVdJGfv_d","txs":[{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000},"sigs":["fbPsFreXByjy0g0y0WQLUDT2KqyiBIC2RbMs2HWU9VNrI4GG1GJMj-9j_Nf0QuMXVvUXEIg3ksOOlSBl30XA3QAgiCJt"],"id":"aLokjgaVT95weTdJmhe2T1VjnvqfqaDNx7JHtRuo8TAsHAps"}]}`, string(jsondata))

	blk.Gas++
	assert.ErrorContains(blk.SyntacticVerify(),
		"Block.SyntacticVerify: invalid gas, expected 638, got 639")
	blk.Gas--

	blk2 := &Block{}
	assert.NoError(blk2.Unmarshal(cbordata))
	assert.NoError(blk2.SyntacticVerify())
	assert.NoError(blk2.TxsMarshalJSON())

	jsondata2, err := json.Marshal(blk2)
	assert.NoError(err)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(blk.ID, blk2.ID)
	assert.Equal(cbordata, blk2.Bytes())
}
