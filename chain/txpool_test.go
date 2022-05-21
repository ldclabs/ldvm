// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxPoolBasic(t *testing.T) {
	assert := assert.New(t)

	tp := NewTxPool()

	tx := ld.MustNewTestTx(util.Signer1, ld.TypeTransfer, &constants.GenesisAccount, nil)
	assert.Equal(0, tp.Len())
	assert.False(tp.Has(tx.ID))
	assert.Nil(tp.Get(tx.ID))

	tp.Add(tx)
	assert.Equal(1, tp.Len())
	assert.True(tp.Has(tx.ID))
	ntx := tp.Get(tx.ID)
	assert.NotNil(ntx)
	assert.Equal(tx.ID, ntx.ID())
	assert.Equal("Unknown", ntx.Status())
	assert.Equal(fmt.Sprintf("%p", tx), fmt.Sprintf("%p", ntx.LD()))

	tp.Add(tx)
	assert.Equal(1, tp.Len(), "should not be added repeatedly")

	tp.Reject(tx)
	assert.Equal(0, tp.Len())
	assert.True(tp.Has(tx.ID))
	ntx = tp.Get(tx.ID)
	assert.NotNil(ntx)
	assert.Equal(tx.ID, ntx.ID())
	assert.Equal("Rejected", ntx.Status())
	assert.Equal(fmt.Sprintf("%p", tx), fmt.Sprintf("%p", ntx.LD()))

	tp.Add(tx)
	assert.Equal(0, tp.Len(), "should not be added after rejected")
	tp.Add(ld.MustNewTestTx(util.Signer1, ld.TypeTest, &constants.GenesisAccount, nil))
	assert.Equal(0, len(tp.txQueue), "TypeTest tx should not be added")

	txs := tp.PopTxsBySize(1000000, 1000, 12345)
	assert.Equal(ld.Txs{}, txs, "no valid tx (1 rejected)")
}

func TestTxPoolRemove(t *testing.T) {
	assert := assert.New(t)

	tp := NewTxPool()
	tx0 := ld.MustNewTestTx(util.Signer1, ld.TypeTransfer, &constants.GenesisAccount, nil)
	tx1 := ld.MustNewTestTx(util.Signer1, ld.TypeTransfer, &constants.GenesisAccount, nil)
	tx2 := ld.MustNewTestTx(util.Signer1, ld.TypeTransfer, &constants.GenesisAccount, nil)
	tx3 := ld.MustNewTestTx(util.Signer1, ld.TypeTransfer, &constants.GenesisAccount, nil)

	tp.Add(tx0, tx1, tx2, tx3)
	assert.Equal(4, tp.Len())
	assert.Equal(tx0.ID, tp.txQueue[0].ID)
	assert.Equal(tx1.ID, tp.txQueue[1].ID)
	assert.Equal(tx2.ID, tp.txQueue[2].ID)
	assert.Equal(tx3.ID, tp.txQueue[3].ID)
	quePtr := fmt.Sprintf("%p", tp.txQueue)

	tp.Remove(tx1.ID)
	assert.False(tp.Has(tx1.ID))
	assert.Equal(3, tp.Len())
	assert.Equal(tx0.ID, tp.txQueue[0].ID)
	assert.Equal(tx2.ID, tp.txQueue[1].ID)
	assert.Equal(tx3.ID, tp.txQueue[2].ID)
	assert.Equal(quePtr, fmt.Sprintf("%p", tp.txQueue), "should not change the underlying array")

	tp.Remove(tx0.ID)
	assert.False(tp.Has(tx0.ID))
	assert.Equal(2, tp.Len())
	assert.Equal(tx2.ID, tp.txQueue[0].ID)
	assert.Equal(tx3.ID, tp.txQueue[1].ID)
	assert.Equal(quePtr, fmt.Sprintf("%p", tp.txQueue), "should not change the underlying array")

	tp.Add(tx0, tx1, tx2, tx3)
	assert.Equal(4, tp.Len())
	assert.Equal(tx2.ID, tp.txQueue[0].ID)
	assert.Equal(tx3.ID, tp.txQueue[1].ID)
	assert.Equal(tx0.ID, tp.txQueue[2].ID)
	assert.Equal(tx1.ID, tp.txQueue[3].ID)
	assert.Equal(quePtr, fmt.Sprintf("%p", tp.txQueue), "should not change the underlying array")

	wg := new(sync.WaitGroup)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			tp.Remove(tx1.ID)
			wg.Done()
		}()
	}
	wg.Wait()
	assert.False(tp.Has(tx1.ID))
	assert.Equal(3, tp.Len())
	assert.Equal(tx2.ID, tp.txQueue[0].ID)
	assert.Equal(tx3.ID, tp.txQueue[1].ID)
	assert.Equal(tx0.ID, tp.txQueue[2].ID)
	assert.Equal(quePtr, fmt.Sprintf("%p", tp.txQueue), "should not change the underlying array")
}

func TestTxPoolPopTxsBySize(t *testing.T) {
	assert := assert.New(t)

	tp := NewTxPool()
	to := util.Signer2.Address()
	s0 := util.NewSigner()
	s1 := util.NewSigner()
	s2 := util.NewSigner()
	s3 := util.NewSigner()

	stx0 := ld.MustNewTestTx(s0, ld.TypeTest, nil, nil)
	stx1 := ld.MustNewTestTx(s0, ld.TypeTransfer, &to, nil)
	stx2 := ld.MustNewTestTx(s1, ld.TypeTransfer, &to, nil)
	stx3 := ld.MustNewTestTx(s1, ld.TypeTransfer, &to, ld.GenJSONData(2500))
	btx, err := ld.NewBatchTx(stx0, stx1, stx2, stx3)
	assert.NoError(err)
	assert.Equal(stx3.ID, btx.ID)
	assert.Equal(stx3.RequiredGas(1000), btx.RequiredGas(1000))

	tx0 := ld.MustNewTestTx(s1, ld.TypeTransfer, &to, nil)
	tx1 := ld.MustNewTestTx(s2, ld.TypeTransfer, &to, ld.GenJSONData(1000))
	tx2 := ld.MustNewTestTx(s3, ld.TypeTransfer, &to, ld.GenJSONData(2000))
	tx3 := ld.MustNewTestTx(s0, ld.TypeTransfer, &to, ld.GenJSONData(3000))
	tp.Add(tx0, tx1, tx2, tx3, btx)
	assert.Equal(7, tp.Len())
	assert.Equal(5, len(tp.txQueue))
	assert.Equal(tx0.ID, tp.txQueue[0].ID)
	assert.Equal(tx1.ID, tp.txQueue[1].ID)
	assert.Equal(tx2.ID, tp.txQueue[2].ID)
	assert.Equal(tx3.ID, tp.txQueue[3].ID)
	assert.Equal(btx.ID, tp.txQueue[4].ID)

	// fmt.Println(tx0.BytesSize()) // 193
	// fmt.Println(tx1.BytesSize()) // 1198
	// fmt.Println(tx2.BytesSize()) // 2198
	// fmt.Println(tx3.BytesSize()) // 3198
	// fmt.Println(btx.BytesSize()) // 3084

	txs := tp.PopTxsBySize(5500, 500, 0)
	// tx2, btx, tx3, tx0, tx1
	assert.Equal(3, tp.Len())
	assert.Equal(3, len(tp.txQueue))
	assert.Equal(2, len(txs))
	assert.Equal(tx2.ID, txs[0].ID)
	assert.Equal(btx.ID, txs[1].ID)

	tx4 := ld.MustNewTestTx(util.Signer1, ld.TypeTransfer, &to, ld.GenJSONData(4000))
	tp.Add(tx4)
	assert.Equal(4, tp.Len())
	assert.Equal(4, len(tp.txQueue))
	txs = tp.PopTxsBySize(5500, 500, 0)
	// tx4, tx3, tx1, tx0
	assert.Equal(3, tp.Len())
	assert.Equal(3, len(tp.txQueue))
	assert.Equal(1, len(txs))
	assert.Equal(tx4.ID, txs[0].ID)

	txs = tp.PopTxsBySize(5500, 500, 0)
	// tx3, tx1, tx0
	assert.Equal(0, tp.Len())
	assert.Equal(0, len(tp.txQueue))
	assert.Equal(3, len(txs))
	assert.Equal(tx3.ID, txs[0].ID)
	assert.Equal(tx1.ID, txs[1].ID)
	assert.Equal(tx0.ID, txs[2].ID)
}
