// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"testing"

	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestSecp256k1Signer(t *testing.T) {
	assert := assert.New(t)

	s1, err := NewSecp256k1()
	assert.NoError(err)

	s2, err := Secp256k1From(s1.PrivateSeed())
	assert.NoError(err)

	assert.Equal(Secp256k1, s1.Kind())
	assert.Equal(Secp256k1, s2.Kind())
	assert.Equal(Secp256k1, s1.Key().Kind())
	assert.True(s1.Key().Equal(s2.Key()))
	assert.Equal(s1.PrivateSeed(), s2.PrivateSeed())

	msg := []byte("hello")
	sig1, err := s1.SignData(msg)
	assert.NoError(err)
	assert.Equal(Secp256k1, sig1.Kind())

	sig2, err := s2.SignData(msg)
	assert.NoError(err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s1.SignHash(util.Sum256(msg))
	assert.NoError(err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s2.SignHash(util.Sum256(msg))
	assert.NoError(err)
	assert.True(sig1.Equal(sig2))
}

func TestEd25519Signer(t *testing.T) {
	assert := assert.New(t)

	s1, err := NewEd25519()
	assert.NoError(err)

	s2, err := Ed25519From(s1.PrivateSeed())
	assert.NoError(err)

	assert.Equal(Ed25519, s1.Kind())
	assert.Equal(Ed25519, s2.Kind())
	assert.Equal(Ed25519, s1.Key().Kind())
	assert.True(s1.Key().Equal(s2.Key()))
	assert.Equal(s1.PrivateSeed(), s2.PrivateSeed())

	msg := []byte("hello")
	sig1, err := s1.SignData(msg)
	assert.NoError(err)
	assert.Equal(Ed25519, sig1.Kind())

	sig2, err := s2.SignData(msg)
	assert.NoError(err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s1.SignHash(util.Sum256(msg))
	assert.NoError(err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s2.SignHash(util.Sum256(msg))
	assert.NoError(err)
	assert.True(sig1.Equal(sig2))
}