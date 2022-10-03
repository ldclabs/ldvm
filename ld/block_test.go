// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestBlock(t *testing.T) {
	assert := assert.New(t)

	var blk *Block
	assert.ErrorContains(blk.SyntacticVerify(), "nil pointer")

	blk = &Block{}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid state 11111111111111111111111111111111LpoYY")

	blk = &Block{State: ids.ID{1, 2, 3}, Timestamp: uint64(time.Now().Unix() + 20)}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid timestamp")

	blk = &Block{State: ids.ID{1, 2, 3}, GasPrice: 1}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid gasPrice")

	blk = &Block{State: ids.ID{1, 2, 3}, GasPrice: 100, GasRebateRate: 1001}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid gasRebateRate")

	blk = &Block{State: ids.ID{1, 2, 3}, GasPrice: 100, Miner: util.StakeSymbol{1, 2, 3}}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid miner address")

	blk = &Block{State: ids.ID{1, 2, 3}, GasPrice: 100}
	assert.ErrorContains(blk.SyntacticVerify(), "no txs")

	blk = &Block{State: ids.ID{1, 2, 3}, GasPrice: 100, Txs: make([]*Transaction, 1),
		Validators: []util.StakeSymbol{{1, 2, 3}}}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid validator address")

	blk = &Block{State: ids.ID{1, 2, 3}, GasPrice: 100, Txs: make([]*Transaction, 1)}
	assert.ErrorContains(blk.SyntacticVerify(), "Transaction.SyntacticVerify error: nil pointer")

	to := util.Signer2.Address()
	tx := &Transaction{
		Tx: TxData{
			Type:      TypeTransfer,
			ChainID:   gChainID,
			Nonce:     1,
			GasTip:    0,
			GasFeeCap: 1000,
			From:      util.Signer1.Address(),
			To:        &to,
			Amount:    big.NewInt(1_000_000),
		},
	}
	assert.NoError(tx.SignWith(util.Signer1))
	assert.NoError(tx.SyntacticVerify())

	blk = &Block{
		State:         ids.ID{1, 2, 3},
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
	assert.Equal(`{"parent":"11111111111111111111111111111111LpoYY","height":0,"timestamp":0,"state":"SkB7qHwfMsyF2PgrjhMvtFxJKhuR5ZfVoW9VATWRV4P9jV7J","gas":638,"gasPrice":1000,"gasRebateRate":200,"miner":"","validators":null,"pChainHeight":0,"id":"qGzUDBq64yjXf9hSdAhe3CsW1RrNhogRR7xGummZfyHnKQQHm","txs":[{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000},"sigs":["7db3ec16b7970728f2d20d32d1640b5034f62aaca20480b645b32cd87594f5536b238186d4624c8fef63fcd7f442e31756f51710883792c38e952065df45c0dd00"],"id":"o87bat4Z2dnH1NKyzD3yMQYaBJKBwQZTmnhQHn4qQFNwDNe5T"}]}`, string(jsondata))

	blk.Gas++
	assert.ErrorContains(blk.SyntacticVerify(),
		"Block.SyntacticVerify error: invalid gas, expected 638, got 639")
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
