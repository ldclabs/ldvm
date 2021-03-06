// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package db

import (
	"math/big"
	"testing"

	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestPrefixDB(t *testing.T) {
	assert := assert.New(t)

	dbp1 := NewPrefixDB(memdb.New(), nil, 100)
	dbp2 := dbp1.With([]byte("LDVM"))
	dbp3 := dbp2.With([]byte("TEST"))

	ok, err := dbp3.Has([]byte("k1"))
	assert.NoError(err)
	assert.False(ok)

	assert.NoError(dbp3.Put([]byte("k1"), []byte("v1")))
	ok, err = dbp3.Has([]byte("k1"))
	assert.NoError(err)
	assert.True(ok)

	v, err := dbp3.Get([]byte("k1"))
	assert.NoError(err)
	assert.Equal([]byte("v1"), v)

	v, err = dbp2.Get([]byte("TESTk1"))
	assert.NoError(err)
	assert.Equal([]byte("v1"), v)

	v, err = dbp1.Get([]byte("LDVMTESTk1"))
	assert.NoError(err)
	assert.Equal([]byte("v1"), v)

	assert.NoError(dbp3.Delete([]byte("k1")))
	ok, err = dbp3.Has([]byte("k1"))
	assert.NoError(err)
	assert.False(ok)

	ok, err = dbp2.Has([]byte("TESTk1"))
	assert.NoError(err)
	assert.False(ok)

	ok, err = dbp1.Has([]byte("LDVMTESTk1"))
	assert.NoError(err)
	assert.False(ok)

	cc := NewCacher(100, 1, func() Objecter { return new(ld.Transaction) })
	to := util.Signer2.Address()
	txData := &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
		Nonce:     1,
		GasTip:    0,
		GasFeeCap: 1000,
		From:      util.Signer1.Address(),
		To:        &to,
		Amount:    big.NewInt(1_000_000),
	}
	sig1, err := util.Signer1.Sign(txData.UnsignedBytes())
	assert.NoError(err)
	txData.Signatures = append(txData.Signatures, sig1)
	tx := txData.ToTransaction()
	assert.NoError(tx.SyntacticVerify())

	_, ok = cc.GetObject(tx.ID[:])
	assert.False(ok)

	assert.NoError(dbp3.Put(tx.ID[:], tx.Bytes()))
	obj, err := dbp3.LoadObject(tx.ID[:], cc)
	assert.NoError(err)
	tx2 := obj.(*ld.Transaction)
	assert.Equal(tx.ID, tx2.ID)
	assert.Equal(tx.Bytes(), tx2.Bytes())

	obj, ok = cc.GetObject(tx.ID[:])
	assert.True(ok)
	tx3 := obj.(*ld.Transaction)
	assert.Equal(tx.Bytes(), tx3.Bytes())
}

func FuzzPrefixDB(f *testing.F) {
	for _, seed := range [][]byte{
		{}, {0}, {9}, {0xa}, {0xf}, {1, 2, 3, 4}, {'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n'},
	} {
		f.Add(seed)
	}
	dbp1 := NewPrefixDB(memdb.New(), nil, 100)
	dbp2 := dbp1.With([]byte("LDVM"))
	f.Fuzz(func(t *testing.T, in []byte) {
		if len(in) > 96 {
			t.Skip()
		}

		assert := assert.New(t)
		assert.NoError(dbp2.Put(in, in))
		v, err := dbp2.Get(in)
		assert.NoError(err)
		assert.Equal(in, v)
		k := make([]byte, 4+len(in))
		copy(k, []byte("LDVM"))
		copy(k[4:], in)
		v, err = dbp1.Get(k)
		assert.NoError(err)
		assert.Equal(in, v)
	})
}
