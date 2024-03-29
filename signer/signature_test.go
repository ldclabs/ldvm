// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ldclabs/ldvm/util/encoding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilSig(t *testing.T) {
	assert := assert.New(t)

	var sig Sig

	assert.Nil(sig)
	assert.True(sig == nil)
	assert.Equal(Unknown, sig.Kind())
	assert.ErrorContains(sig.Valid(), "unknown sig p__G-A")
	assert.ErrorContains((&sig).Valid(), "unknown sig p__G-A")

	assert.Nil(sig.Bytes())
	assert.Equal("p__G-A", sig.String())
	assert.Equal("p__G-A", sig.GoString())
	assert.True(sig.Equal(sig))

	b, err := sig.MarshalText()
	require.NoError(t, err)
	assert.Equal("p__G-A", string(b))
	assert.NoError(sig.UnmarshalText(b))
	assert.Nil(sig)

	b, err = sig.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(`"p__G-A"`, string(b))
	assert.NoError(sig.UnmarshalJSON(b))
	assert.Nil(sig)

	b, err = sig.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR(nil), b)
	assert.NoError(sig.UnmarshalCBOR(b))
	assert.Nil(sig)
	assert.Nil(sig.Clone())

	msg := encoding.Sum256([]byte("hello"))
	assert.Equal(-1, sig.FindKey(msg))
	assert.Equal(-1, sig.FindKey(msg, Signer1.Key()))
}

func TestEmptySig(t *testing.T) {
	assert := assert.New(t)

	sig := Sig{}
	var sig2 Sig

	require.NotNil(t, sig)
	assert.Equal(Unknown, sig.Kind())
	assert.ErrorContains(sig.Valid(), "unknown sig p__G-A")
	assert.ErrorContains((&sig).Valid(), "unknown sig p__G-A")

	assert.Equal([]byte{}, sig.Bytes())
	assert.Equal("p__G-A", sig.String())
	assert.Equal("p__G-A", sig.GoString())
	assert.True(sig.Equal(sig))
	assert.False(sig.Equal(sig2))

	b, err := sig.MarshalText()
	require.NoError(t, err)
	assert.Equal("p__G-A", string(b))
	assert.NoError(sig.UnmarshalText(b))
	assert.Equal([]byte{}, sig.Bytes())

	b, err = sig.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(`"p__G-A"`, string(b))
	assert.NoError(sig.UnmarshalJSON(b))
	assert.Equal([]byte{}, sig.Bytes())
	assert.Equal([]byte{}, sig.Clone().Bytes())

	b, err = sig.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR([]byte{}), b)
	assert.NoError(sig.UnmarshalCBOR(b))
	assert.Equal([]byte{}, sig.Bytes())
	assert.Equal([]byte{}, sig.Clone().Bytes())

	msg := encoding.Sum256([]byte("hello"))
	assert.Equal(-1, sig.FindKey(msg))
	assert.Equal(-1, sig.FindKey(msg, Signer1.Key()))
}

func TestSecp256k1Sig(t *testing.T) {
	assert := assert.New(t)

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer1.SignHash(msg)
	require.NoError(t, err)

	sig2 := Sig{}
	data := sig.Bytes()
	sigStr := "CFK1lA5EyeTrgWgkYJbmdfIaJWaUfMScTh3BWAupPZsjqgspDQ75bqwNLTzswotIvPGE2mbJj8wh71W5cV8aEgBKfqjO"

	require.NotNil(t, sig)
	assert.Equal(Secp256k1, sig.Kind())

	assert.Equal(data, sig.Bytes())
	assert.Equal(sigStr, sig.String())
	assert.Equal(sigStr, sig.GoString())
	assert.True(sig.Equal(sig))
	assert.False(sig.Equal(sig2))

	b, err := sig.MarshalText()
	require.NoError(t, err)
	assert.Equal(sigStr, string(b))
	assert.NoError(sig2.UnmarshalText(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(strconv.Quote(sigStr), string(b))
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalJSON(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR(data), b)
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalCBOR(b))
	assert.Equal(data, sig2.Bytes())
	assert.Equal(data, sig2.Clone().Bytes())

	assert.Equal(-1, sig.FindKey(msg))
	assert.Equal(0, sig.FindKey(msg, Signer1.Key()))
	assert.Equal(1, sig.FindKey(msg, Signer2.Key(), Signer1.Key()))
	assert.Equal(2, sig.FindKey(msg, Signer3.Key(), Signer2.Key(), Signer1.Key()))
}

func TestEd25519Sig(t *testing.T) {
	assert := assert.New(t)

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer3.SignHash(msg)
	require.NoError(t, err)

	sig2 := Sig{}
	data := sig.Bytes()
	sigStr := "6Uik1OFvj2ULuT0KTBwF3u62Fw-i0xS0-ftzWwUlO7ylS35taKgmv5psNeiUiTN93BqPrwz2X_HszbZJ6hxwBBd-tWI"

	require.NotNil(t, sig)
	assert.Equal(Ed25519, sig.Kind())

	assert.Equal(data, sig.Bytes())
	assert.Equal(sigStr, sig.String())
	assert.Equal(sigStr, sig.GoString())
	assert.True(sig.Equal(sig))
	assert.False(sig.Equal(sig2))

	b, err := sig.MarshalText()
	require.NoError(t, err)
	assert.Equal(sigStr, string(b))
	assert.NoError(sig2.UnmarshalText(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(strconv.Quote(sigStr), string(b))
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalJSON(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR(data), b)
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalCBOR(b))
	assert.Equal(data, sig2.Bytes())
	assert.Equal(data, sig2.Clone().Bytes())

	assert.Equal(-1, sig.FindKey(msg))
	assert.Equal(0, sig.FindKey(msg, Signer3.Key()))
	assert.Equal(1, sig.FindKey(msg, Signer2.Key(), Signer3.Key()))
	assert.Equal(2, sig.FindKey(msg, Signer1.Key(), Signer2.Key(), Signer3.Key()))
}

func TestBLS12381Sig(t *testing.T) {
	assert := assert.New(t)

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer4.SignHash(msg)
	require.NoError(t, err)

	sig2 := Sig{}
	data := sig.Bytes()
	sigStr := "kUu8w4TvwF5jgsccH72NmbBcYlE0gTQB4dBuDhwJwThjN_GI5loddLnTptMGKtaIGFLGiiUlQCiwDoQw1Hb1lcP8HEEvREQfkuyPuIc7BCBdvBQsFbBdIiFC9_ABP2mOZtaGyg"

	require.NotNil(t, sig)
	assert.Equal(BLS12381, sig.Kind())

	assert.Equal(data, sig.Bytes())
	assert.Equal(sigStr, sig.String())
	assert.Equal(sigStr, sig.GoString())
	assert.True(sig.Equal(sig))
	assert.False(sig.Equal(sig2))

	b, err := sig.MarshalText()
	require.NoError(t, err)
	assert.Equal(sigStr, string(b))
	assert.NoError(sig2.UnmarshalText(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(strconv.Quote(sigStr), string(b))
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalJSON(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR(data), b)
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalCBOR(b))
	assert.Equal(data, sig2.Bytes())
	assert.Equal(data, sig2.Clone().Bytes())

	assert.Equal(-1, sig.FindKey(msg))
	assert.Equal(0, sig.FindKey(msg, Signer4.Key()))
	assert.Equal(1, sig.FindKey(msg, Signer2.Key(), Signer4.Key()))
	assert.Equal(2, sig.FindKey(msg, Signer1.Key(), Signer2.Key(), Signer4.Key()))
}

func TestSigs(t *testing.T) {
	assert := assert.New(t)

	var sigs Sigs
	assert.NoError(sigs.Valid())

	sigs = Sigs{}
	assert.NoError(sigs.Valid())

	msg := encoding.Sum256([]byte("LDC Labs"))
	sig1, err := Signer1.SignHash(msg)
	require.NoError(t, err)
	sig2, err := Signer2.SignHash(msg)
	require.NoError(t, err)
	sig3, err := Signer4.SignHash(msg)
	require.NoError(t, err)

	assert.NoError(Sigs{sig1, sig2, sig3}.Valid())
	assert.ErrorContains(Sigs{sig1, sig2, sig3, sig3}.Valid(), "duplicate sig gMM7IYZaMz6Zq-yetgri4HB2wp7IlXZNVMpP1eoqtLC3bZrjN4pAx0UhxwaLYGPFEvdhGyLs3_ZW6wdAhyVBV8xpEVfvnUxj5FA7thIeBRK_ZCTacGoJ0rWpmO8C10QXefWbIg")

	assert.ErrorContains(Sigs{sig1, sig2, sig3, Sig{}}.Valid(), "unknown sig p__G-A")
}

func TestSigInStruct(t *testing.T) {
	assert := assert.New(t)

	type T1 struct {
		Sig *Sig `cbor:"s,omitempty"`
	}

	t1 := &T1{Sig: nil}
	d, err := encoding.MarshalCBOR(t1)
	require.NoError(t, err)
	assert.Equal("a0", fmt.Sprintf("%x", d))

	var t1b T1
	assert.NoError(encoding.UnmarshalCBOR(d, &t1b))

	var sig Sig
	t1 = &T1{Sig: &sig}
	d, err = encoding.MarshalCBOR(t1)
	require.NoError(t, err)
	assert.Equal("a16173f6", fmt.Sprintf("%x", d))
	assert.NoError(encoding.UnmarshalCBOR(d, &t1b))
	assert.True(t1b.Sig == nil)
	assert.Equal("a0", fmt.Sprintf("%x", encoding.MustMarshalCBOR(t1b))) // fix?

	t1 = &T1{Sig: &Sig{}}
	d, err = encoding.MarshalCBOR(t1)
	require.NoError(t, err)
	assert.Equal("a1617340", fmt.Sprintf("%x", d))
	assert.NoError(encoding.UnmarshalCBOR(d, &t1b))
	assert.True(t1b.Sig != nil)
	assert.Equal("a1617340", fmt.Sprintf("%x", encoding.MustMarshalCBOR(t1b)))

	type T2 struct {
		Sig Sig `cbor:"s,omitempty"`
	}

	t2 := &T2{}
	d, err = encoding.MarshalCBOR(t2)
	require.NoError(t, err)
	assert.Equal("a16173f6", fmt.Sprintf("%x", d))

	var t2b T2
	assert.NoError(encoding.UnmarshalCBOR(d, &t2b))

	t2 = &T2{Sig: sig}
	d, err = encoding.MarshalCBOR(t2)
	require.NoError(t, err)
	assert.Equal("a16173f6", fmt.Sprintf("%x", d))
	assert.NoError(encoding.UnmarshalCBOR(d, &t2b))
	assert.True(t2b.Sig == nil)
	assert.Equal("a16173f6", fmt.Sprintf("%x", encoding.MustMarshalCBOR(t2b)))

	t2 = &T2{Sig: Sig{}}
	d, err = encoding.MarshalCBOR(t2)
	require.NoError(t, err)
	assert.Equal("a1617340", fmt.Sprintf("%x", d))
	assert.NoError(encoding.UnmarshalCBOR(d, &t2b))
	assert.True(t2b.Sig != nil)
	assert.Equal("a1617340", fmt.Sprintf("%x", encoding.MustMarshalCBOR(t2b)))
}
