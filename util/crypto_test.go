// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	assert := assert.New(t)

	data := []byte("Hello 👋")
	sig1, err := Signer1.Sign(data)
	assert.Nil(err)

	sig2, err := Signer2.Sign(data)
	assert.Nil(err)

	assert.NotNil(sig1, sig2)

	addr1, err := DeriveSigner(data, sig1[:])
	assert.Nil(err)
	assert.Equal(Signer1.Address(), addr1)

	addrs, err := DeriveSigners(data, []Signature{sig1, sig2})
	assert.Nil(err)
	assert.Equal(Signer1.Address(), addrs[0])
	assert.Equal(Signer2.Address(), addrs[1])

	sigd, err := json.Marshal(sig1)
	assert.Nil(err)
	var sig Signature
	assert.Nil(json.Unmarshal(sigd, &sig))
	assert.Equal(sig1, sig)

	sigd, err = cbor.Marshal(sig2)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(sigd, &sig))
	assert.Equal(sig2, sig)

	sigs, err := SignaturesFromStrings([]string{sig1.String(), sig2.String()})
	assert.Nil(err)
	assert.Equal(sig1, sigs[0])
	assert.Equal(sig2, sigs[1])
}

func TestSatisfySigning(t *testing.T) {
	assert := assert.New(t)

	ks := make([]EthID, 4)
	for i := range ks {
		rand.Read(ks[i][:])
	}

	assert.True(SatisfySigning(0, []EthID{}, []EthID{}, true))
	assert.False(SatisfySigning(0, []EthID{}, []EthID{}, false))
	assert.False(SatisfySigningPlus(0, []EthID{}, []EthID{}))

	assert.True(SatisfySigning(1, []EthID{}, []EthID{}, true))
	assert.False(SatisfySigning(1, []EthID{}, []EthID{}, false))
	assert.False(SatisfySigningPlus(1, []EthID{}, []EthID{}))

	assert.True(SatisfySigning(0, []EthID{ks[0]}, []EthID{}, true))
	assert.False(SatisfySigning(0, []EthID{ks[0]}, []EthID{}, false))
	assert.False(SatisfySigningPlus(0, []EthID{ks[0]}, []EthID{}))

	assert.True(SatisfySigning(1, []EthID{ks[0]}, []EthID{ks[0]}, true))
	assert.True(SatisfySigning(1, []EthID{ks[0]}, []EthID{ks[0]}, false))
	assert.True(SatisfySigningPlus(1, []EthID{ks[0]}, []EthID{ks[0]}))
	assert.False(SatisfySigningPlus(1, []EthID{ks[0], ks[1]}, []EthID{ks[0]}))
	assert.True(SatisfySigningPlus(1, []EthID{ks[0], ks[1]}, []EthID{ks[0], ks[1]}))

	assert.True(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2]}, []EthID{ks[0], ks[2]}, false))
	assert.False(SatisfySigningPlus(2, []EthID{ks[0], ks[1], ks[2]}, []EthID{ks[0], ks[2]}))
	assert.True(SatisfySigningPlus(2, []EthID{ks[0], ks[1], ks[2]}, []EthID{ks[0], ks[1], ks[2]}))
	assert.True(SatisfySigningPlus(3, []EthID{ks[0], ks[1], ks[2]}, []EthID{ks[0], ks[1], ks[2]}))

	assert.False(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2]}, []EthID{ks[0], ks[3]}, false))
	assert.False(SatisfySigningPlus(2, []EthID{ks[0], ks[1], ks[2]}, []EthID{ks[0], ks[3]}))
	assert.False(SatisfySigningPlus(2, []EthID{ks[0], ks[1], ks[2]}, []EthID{ks[0], ks[1], ks[3]}))
	assert.False(SatisfySigningPlus(3, []EthID{ks[0], ks[1], ks[2]}, []EthID{ks[0], ks[1], ks[3]}))

	assert.False(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0]}, false))
	assert.True(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0], ks[1]}, false))
	assert.True(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0], ks[2]}, false))
	assert.True(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0], ks[3]}, false))
	assert.True(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[1], ks[2]}, false))
	assert.True(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[1], ks[3]}, false))
	assert.True(SatisfySigning(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[2], ks[3]}, false))

	assert.False(SatisfySigningPlus(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0], ks[1]}))
	assert.True(SatisfySigningPlus(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0], ks[1], ks[2]}))
	assert.True(SatisfySigningPlus(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0], ks[1], ks[3]}))
	assert.True(SatisfySigningPlus(2, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[1], ks[2], ks[3]}))
	assert.False(SatisfySigningPlus(3, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[1], ks[2], ks[3]}))
	assert.True(SatisfySigningPlus(3, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0], ks[1], ks[2], ks[3]}))
	assert.True(SatisfySigningPlus(4, []EthID{ks[0], ks[1], ks[2], ks[3]}, []EthID{ks[0], ks[1], ks[2], ks[3]}))
}
