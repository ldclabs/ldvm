// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(err)
	assert.Equal("p__G-A", string(b))
	assert.NoError(sig.UnmarshalText(b))
	assert.Nil(sig)

	b, err = sig.MarshalJSON()
	assert.NoError(err)
	assert.Equal(`"p__G-A"`, string(b))
	assert.NoError(sig.UnmarshalJSON(b))
	assert.Nil(sig)

	b, err = sig.MarshalCBOR()
	assert.NoError(err)
	assert.Equal(util.MustMarshalCBOR(nil), b)
	assert.NoError(sig.UnmarshalCBOR(b))
	assert.Nil(sig)
	assert.Nil(sig.Clone())

	msg := util.Sum256([]byte("hello"))
	assert.Equal(-1, sig.FindKey(msg))
	assert.Equal(-1, sig.FindKey(msg, Signer1.Key()))
}

func TestEmptySig(t *testing.T) {
	assert := assert.New(t)

	sig := Sig{}
	var sig2 Sig

	assert.NotNil(sig)
	assert.True(sig != nil)
	assert.Equal(Unknown, sig.Kind())
	assert.ErrorContains(sig.Valid(), "unknown sig p__G-A")
	assert.ErrorContains((&sig).Valid(), "unknown sig p__G-A")

	assert.Equal([]byte{}, sig.Bytes())
	assert.Equal("p__G-A", sig.String())
	assert.Equal("p__G-A", sig.GoString())
	assert.True(sig.Equal(sig))
	assert.False(sig.Equal(sig2))

	b, err := sig.MarshalText()
	assert.NoError(err)
	assert.Equal("p__G-A", string(b))
	assert.NoError(sig.UnmarshalText(b))
	assert.Equal([]byte{}, sig.Bytes())

	b, err = sig.MarshalJSON()
	assert.NoError(err)
	assert.Equal(`"p__G-A"`, string(b))
	assert.NoError(sig.UnmarshalJSON(b))
	assert.Equal([]byte{}, sig.Bytes())
	assert.Equal([]byte{}, sig.Clone().Bytes())

	b, err = sig.MarshalCBOR()
	assert.NoError(err)
	assert.Equal(util.MustMarshalCBOR([]byte{}), b)
	assert.NoError(sig.UnmarshalCBOR(b))
	assert.Equal([]byte{}, sig.Bytes())
	assert.Equal([]byte{}, sig.Clone().Bytes())

	msg := util.Sum256([]byte("hello"))
	assert.Equal(-1, sig.FindKey(msg))
	assert.Equal(-1, sig.FindKey(msg, Signer1.Key()))
}

func TestSecp256k1Sig(t *testing.T) {
	assert := assert.New(t)

	msg := util.Sum256([]byte("hello"))
	sig, err := Signer1.SignHash(msg)
	assert.NoError(err)

	sig2 := Sig{}
	data := sig.Bytes()
	sigStr := "CFK1lA5EyeTrgWgkYJbmdfIaJWaUfMScTh3BWAupPZsjqgspDQ75bqwNLTzswotIvPGE2mbJj8wh71W5cV8aEgBKfqjO"

	assert.NotNil(sig)
	assert.True(sig != nil)
	assert.Equal(Secp256k1, sig.Kind())

	assert.Equal(data, sig.Bytes())
	assert.Equal(sigStr, sig.String())
	assert.Equal(sigStr, sig.GoString())
	assert.True(sig.Equal(sig))
	assert.False(sig.Equal(sig2))

	b, err := sig.MarshalText()
	assert.NoError(err)
	assert.Equal(sigStr, string(b))
	assert.NoError(sig2.UnmarshalText(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalJSON()
	assert.NoError(err)
	assert.Equal(strconv.Quote(sigStr), string(b))
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalJSON(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalCBOR()
	assert.NoError(err)
	assert.Equal(util.MustMarshalCBOR(data), b)
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

	msg := util.Sum256([]byte("hello"))
	sig, err := Signer3.SignHash(msg)
	assert.NoError(err)

	sig2 := Sig{}
	data := sig.Bytes()
	sigStr := "6Uik1OFvj2ULuT0KTBwF3u62Fw-i0xS0-ftzWwUlO7ylS35taKgmv5psNeiUiTN93BqPrwz2X_HszbZJ6hxwBBd-tWI"

	assert.NotNil(sig)
	assert.True(sig != nil)
	assert.Equal(Ed25519, sig.Kind())

	assert.Equal(data, sig.Bytes())
	assert.Equal(sigStr, sig.String())
	assert.Equal(sigStr, sig.GoString())
	assert.True(sig.Equal(sig))
	assert.False(sig.Equal(sig2))

	b, err := sig.MarshalText()
	assert.NoError(err)
	assert.Equal(sigStr, string(b))
	assert.NoError(sig2.UnmarshalText(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalJSON()
	assert.NoError(err)
	assert.Equal(strconv.Quote(sigStr), string(b))
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalJSON(b))
	assert.Equal(data, sig2.Bytes())

	b, err = sig.MarshalCBOR()
	assert.NoError(err)
	assert.Equal(util.MustMarshalCBOR(data), b)
	sig2 = Sig{}
	assert.NoError(sig2.UnmarshalCBOR(b))
	assert.Equal(data, sig2.Bytes())
	assert.Equal(data, sig2.Clone().Bytes())

	assert.Equal(-1, sig.FindKey(msg))
	assert.Equal(0, sig.FindKey(msg, Signer3.Key()))
	assert.Equal(1, sig.FindKey(msg, Signer2.Key(), Signer3.Key()))
	assert.Equal(2, sig.FindKey(msg, Signer1.Key(), Signer2.Key(), Signer3.Key()))
}

func TestSigs(t *testing.T) {
	assert := assert.New(t)

	var sigs Sigs
	assert.NoError(sigs.Valid())

	sigs = Sigs{}
	assert.NoError(sigs.Valid())

	msg := util.Sum256([]byte("LDC Labs"))
	sig1, err := Signer1.SignHash(msg)
	assert.NoError(err)
	sig2, err := Signer2.SignHash(msg)
	assert.NoError(err)
	sig3, err := Signer4.SignHash(msg)
	assert.NoError(err)

	assert.NoError(Sigs{sig1, sig2, sig3}.Valid())
	assert.ErrorContains(Sigs{sig1, sig2, sig3, sig3}.Valid(), "duplicate sig xYmDC3UoLgAO0nYZ6zsQuPiZbCkZBFfDBEuq8BSx8zILpnMGZ4WE1VRjrpA5mZuvT7Ga9QmUttEWBn97gtnNDWp08G0")

	assert.ErrorContains(Sigs{sig1, sig2, sig3, Sig{}}.Valid(), "unknown sig p__G-A")
}

func TestSigInStruct(t *testing.T) {
	assert := assert.New(t)

	type T1 struct {
		Sig *Sig `cbor:"s,omitempty"`
	}

	t1 := &T1{Sig: nil}
	d, err := util.MarshalCBOR(t1)
	assert.NoError(err)
	assert.Equal("a0", fmt.Sprintf("%x", d))

	var t1b T1
	assert.NoError(util.UnmarshalCBOR(d, &t1b))

	var sig Sig
	t1 = &T1{Sig: &sig}
	d, err = util.MarshalCBOR(t1)
	assert.NoError(err)
	assert.Equal("a16173f6", fmt.Sprintf("%x", d))
	assert.NoError(util.UnmarshalCBOR(d, &t1b))
	assert.True(t1b.Sig == nil)
	assert.Equal("a0", fmt.Sprintf("%x", util.MustMarshalCBOR(t1b))) // fix?

	t1 = &T1{Sig: &Sig{}}
	d, err = util.MarshalCBOR(t1)
	assert.NoError(err)
	assert.Equal("a1617340", fmt.Sprintf("%x", d))
	assert.NoError(util.UnmarshalCBOR(d, &t1b))
	assert.True(t1b.Sig != nil)
	assert.Equal("a1617340", fmt.Sprintf("%x", util.MustMarshalCBOR(t1b)))

	type T2 struct {
		Sig Sig `cbor:"s,omitempty"`
	}

	t2 := &T2{}
	d, err = util.MarshalCBOR(t2)
	assert.NoError(err)
	assert.Equal("a16173f6", fmt.Sprintf("%x", d))

	var t2b T2
	assert.NoError(util.UnmarshalCBOR(d, &t2b))

	t2 = &T2{Sig: sig}
	d, err = util.MarshalCBOR(t2)
	assert.NoError(err)
	assert.Equal("a16173f6", fmt.Sprintf("%x", d))
	assert.NoError(util.UnmarshalCBOR(d, &t2b))
	assert.True(t2b.Sig == nil)
	assert.Equal("a16173f6", fmt.Sprintf("%x", util.MustMarshalCBOR(t2b)))

	t2 = &T2{Sig: Sig{}}
	d, err = util.MarshalCBOR(t2)
	assert.NoError(err)
	assert.Equal("a1617340", fmt.Sprintf("%x", d))
	assert.NoError(util.UnmarshalCBOR(d, &t2b))
	assert.True(t2b.Sig != nil)
	assert.Equal("a1617340", fmt.Sprintf("%x", util.MustMarshalCBOR(t2b)))
}
