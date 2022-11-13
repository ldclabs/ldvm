// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"testing"

	"github.com/ldclabs/ldvm/util/encoding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecp256k1Signer(t *testing.T) {
	assert := assert.New(t)

	s1, err := NewSecp256k1()
	require.NoError(t, err)

	s2, err := Secp256k1From(s1.PrivateSeed())
	require.NoError(t, err)

	assert.Equal(Secp256k1, s1.Kind())
	assert.Equal(Secp256k1, s2.Kind())
	assert.Equal(Secp256k1, s1.Key().Kind())
	assert.True(s1.Key().Equal(s2.Key()))
	assert.Equal(s1.PrivateSeed(), s2.PrivateSeed())

	msg := []byte("hello")
	sig1, err := s1.SignData(msg)
	require.NoError(t, err)
	assert.Equal(Secp256k1, sig1.Kind())

	sig2, err := s2.SignData(msg)
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s1.SignHash(encoding.Sum256(msg))
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s2.SignHash(encoding.Sum256(msg))
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))
}

func TestEd25519Signer(t *testing.T) {
	assert := assert.New(t)

	s1, err := NewEd25519()
	require.NoError(t, err)

	s2, err := Ed25519From(s1.PrivateSeed())
	require.NoError(t, err)

	assert.Equal(Ed25519, s1.Kind())
	assert.Equal(Ed25519, s2.Kind())
	assert.Equal(Ed25519, s1.Key().Kind())
	assert.True(s1.Key().Equal(s2.Key()))
	assert.Equal(s1.PrivateSeed(), s2.PrivateSeed())

	msg := []byte("hello")
	sig1, err := s1.SignData(msg)
	require.NoError(t, err)
	assert.Equal(Ed25519, sig1.Kind())

	sig2, err := s2.SignData(msg)
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s1.SignHash(encoding.Sum256(msg))
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s2.SignHash(encoding.Sum256(msg))
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))
}

func TestBLS12381Signer(t *testing.T) {
	assert := assert.New(t)

	s1, err := NewBLS12381()
	require.NoError(t, err)

	s2, err := BLS12381From(s1.PrivateSeed())
	require.NoError(t, err)

	assert.Equal(BLS12381, s1.Kind())
	assert.Equal(BLS12381, s2.Kind())
	assert.Equal(BLS12381, s1.Key().Kind())
	assert.True(s1.Key().Equal(s2.Key()))
	assert.Equal(s1.PrivateSeed(), s2.PrivateSeed())

	msg := []byte("hello")
	sig1, err := s1.SignData(msg)
	require.NoError(t, err)
	assert.Equal(BLS12381, sig1.Kind())

	sig2, err := s2.SignData(msg)
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s1.SignHash(encoding.Sum256(msg))
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))

	sig2, err = s2.SignHash(encoding.Sum256(msg))
	require.NoError(t, err)
	assert.True(sig1.Equal(sig2))
}
