// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/encoding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilKey(t *testing.T) {
	assert := assert.New(t)

	var key Key

	assert.Nil(key)
	assert.True(key == nil)
	assert.Equal(Unknown, key.Kind())
	assert.ErrorContains(key.ValidOrEmpty(), "nil key")
	assert.ErrorContains(key.Valid(), "empty key")

	assert.False(key.IsAddress(Signer1.Key().Address()))
	assert.Equal(ids.EmptyAddress, key.Address())
	assert.Nil(key.Bytes())
	assert.Equal("p__G-A", key.String())
	assert.Equal("p__G-A", key.GoString())
	assert.True(key.Equal(key))

	b, err := key.MarshalText()
	require.NoError(t, err)
	assert.Equal("p__G-A", string(b))
	assert.NoError(key.UnmarshalText(b), "empty key")
	assert.Nil(key)

	b, err = key.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(`"p__G-A"`, string(b))
	assert.NoError(key.UnmarshalJSON(b), "empty key")
	assert.Nil(key)

	b, err = key.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR(nil), b)
	assert.NoError(key.UnmarshalCBOR(b), "empty key")
	assert.Nil(key)
	assert.Nil(key.Clone())

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer1.SignHash(msg)
	require.NoError(t, err)
	assert.False(key.Verify(msg, Sigs{sig}))
}

func TestEmptyKey(t *testing.T) {
	assert := assert.New(t)

	key := Key{}
	var key2 Key

	require.NotNil(t, key)
	assert.Equal(Unknown, key.Kind())
	assert.NoError(key.ValidOrEmpty())
	assert.ErrorContains(key.Valid(), "empty key")

	assert.False(key.IsAddress(Signer1.Key().Address()))
	assert.Equal(ids.EmptyAddress, key.Address())
	assert.Equal([]byte{}, key.Bytes())
	assert.Equal("p__G-A", key.String())
	assert.Equal("p__G-A", key.GoString())
	assert.True(key.Equal(key))
	assert.False(key.Equal(key2))

	b, err := key.MarshalText()
	require.NoError(t, err)
	assert.Equal("p__G-A", string(b))
	assert.NoError(key.UnmarshalText(b), "empty key")
	assert.Equal([]byte{}, key.Bytes())

	b, err = key.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(`"p__G-A"`, string(b))
	assert.NoError(key.UnmarshalJSON(b), "empty key")
	assert.Equal([]byte{}, key.Bytes())

	b, err = key.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR([]byte{}), b)
	assert.NoError(key.UnmarshalCBOR(b), "empty key")
	assert.Equal([]byte{}, key.Bytes())
	assert.Equal([]byte{}, key.Clone().Bytes())

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer1.SignHash(msg)
	require.NoError(t, err)
	assert.False(key.Verify(msg, Sigs{sig}))
}

func TestSecp256k1Key(t *testing.T) {
	assert := assert.New(t)

	key := Signer1.Key()
	key2 := Key{}
	data := Signer1.Key().Bytes()
	keyStr := "jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"

	require.NotNil(t, key)
	assert.Equal(Secp256k1, key.Kind())
	assert.NoError(key.Valid())

	assert.True(key.IsAddress(Signer1.Key().Address()))
	assert.False(key.IsAddress(Signer2.Key().Address()))
	assert.Equal(Signer1.Key().Address(), key.Address())
	assert.Equal(data, key.Bytes())
	assert.Equal(keyStr, key.String())
	assert.Equal(keyStr, key.GoString())
	assert.Equal("0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc", key.Address().String())
	assert.Equal("0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc", key.Address().GoString())
	assert.True(key.Equal(key))
	assert.False(key.Equal(key2))

	b, err := key.MarshalText()
	require.NoError(t, err)
	assert.Equal(keyStr, string(b))
	assert.NoError(key2.UnmarshalText(b))
	assert.Equal(data, key2.Bytes())

	b, err = key.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(strconv.Quote(keyStr), string(b))
	key2 = Key{}
	assert.NoError(key2.UnmarshalJSON(b))
	assert.Equal(data, key2.Bytes())

	b, err = key.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR(data), b)
	key2 = Key{}
	assert.NoError(key2.UnmarshalCBOR(b))
	assert.Equal(data, key2.Bytes())
	assert.Equal(data, key2.Clone().Bytes())

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer1.SignHash(msg)
	require.NoError(t, err)
	assert.True(key.Verify(msg, Sigs{sig}))
	assert.True(key2.Verify(msg, Sigs{sig}))

	sig2, err := Signer2.SignHash(msg)
	require.NoError(t, err)
	assert.False(key.Verify(msg, Sigs{sig2}))
	assert.False(key2.Verify(msg, Sigs{sig2}))
	assert.True(key.Verify(msg, Sigs{sig2, sig}))
}

func TestEd25519Key(t *testing.T) {
	assert := assert.New(t)

	key := Signer3.Key()
	key2 := Key{}
	data := key.Bytes()
	keyStr := "OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F"

	require.NotNil(t, key)
	assert.Equal(Ed25519, key.Kind())
	assert.NoError(key.Valid())

	assert.True(key.IsAddress(Signer3.Key().Address()))
	assert.False(key.IsAddress(Signer2.Key().Address()))
	assert.False(key.IsAddress(Signer4.Key().Address()))
	assert.Equal(Signer3.Key().Address(), key.Address())
	assert.Equal(data, key.Bytes())
	assert.Equal(keyStr, key.String())
	assert.Equal(keyStr, key.GoString())
	assert.Equal("0x6962DD0564Fb1f8459624e5b7c5dD9A38b2F990d", key.Address().String())
	assert.Equal("0x6962DD0564Fb1f8459624e5b7c5dD9A38b2F990d", key.Address().GoString())
	assert.True(key.Equal(key))
	assert.False(key.Equal(key2))

	b, err := key.MarshalText()
	require.NoError(t, err)
	assert.Equal(keyStr, string(b))
	assert.NoError(key2.UnmarshalText(b))
	assert.Equal(data, key2.Bytes())

	b, err = key.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(strconv.Quote(keyStr), string(b))
	key2 = Key{}
	assert.NoError(key2.UnmarshalJSON(b))
	assert.Equal(data, key2.Bytes())

	b, err = key.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR(data), b)
	key2 = Key{}
	assert.NoError(key2.UnmarshalCBOR(b))
	assert.Equal(data, key2.Bytes())
	assert.Equal(data, key2.Clone().Bytes())

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer3.SignHash(msg)
	require.NoError(t, err)
	assert.True(key.Verify(msg, Sigs{sig}))
	assert.True(key2.Verify(msg, Sigs{sig}))

	sig2, err := Signer2.SignHash(msg)
	require.NoError(t, err)
	assert.False(key.Verify(msg, Sigs{sig2}))
	assert.False(key2.Verify(msg, Sigs{sig2}))
	assert.True(key.Verify(msg, Sigs{sig2, sig}))

	sig3, err := Signer4.SignHash(msg)
	require.NoError(t, err)
	assert.False(key.Verify(msg, Sigs{sig3}))
	assert.False(key2.Verify(msg, Sigs{sig3}))
	assert.True(key.Verify(msg, Sigs{sig3, sig2, sig}))
}

func TestBLS12381Key(t *testing.T) {
	assert := assert.New(t)

	key := Signer4.Key()
	key2 := Key{}
	data := key.Bytes()
	keyStr := "hJEADz4AlkZ_NSt41-9x5eTaahzNzgMzd0wOBF-B2kJGSpWTCQutstgl0tXrZKQVIsBdNQ"

	require.NotNil(t, key)
	assert.Equal(BLS12381, key.Kind())
	assert.NoError(key.Valid())

	assert.True(key.IsAddress(Signer4.Key().Address()))
	assert.False(key.IsAddress(Signer2.Key().Address()))
	assert.False(key.IsAddress(Signer3.Key().Address()))
	assert.Equal(Signer4.Key().Address(), key.Address())
	assert.Equal(data, key.Bytes())
	assert.Equal(keyStr, key.String())
	assert.Equal(keyStr, key.GoString())
	assert.Equal("0x22C05D35Be1305c33810086d3A4dB598c3E1Cf48", key.Address().String())
	assert.Equal("0x22C05D35Be1305c33810086d3A4dB598c3E1Cf48", key.Address().GoString())
	assert.True(key.Equal(key))
	assert.False(key.Equal(key2))

	b, err := key.MarshalText()
	require.NoError(t, err)
	assert.Equal(keyStr, string(b))
	assert.NoError(key2.UnmarshalText(b))
	assert.Equal(data, key2.Bytes())

	b, err = key.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(strconv.Quote(keyStr), string(b))
	key2 = Key{}
	assert.NoError(key2.UnmarshalJSON(b))
	assert.Equal(data, key2.Bytes())

	b, err = key.MarshalCBOR()
	require.NoError(t, err)
	assert.Equal(encoding.MustMarshalCBOR(data), b)
	key2 = Key{}
	assert.NoError(key2.UnmarshalCBOR(b))
	assert.Equal(data, key2.Bytes())
	assert.Equal(data, key2.Clone().Bytes())

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer4.SignHash(msg)
	require.NoError(t, err)
	assert.True(key.Verify(msg, Sigs{sig}))
	assert.True(key2.Verify(msg, Sigs{sig}))

	sig2, err := Signer2.SignHash(msg)
	require.NoError(t, err)
	assert.False(key.Verify(msg, Sigs{sig2}))
	assert.False(key2.Verify(msg, Sigs{sig2}))
	assert.True(key.Verify(msg, Sigs{sig2, sig}))

	sig3, err := Signer3.SignHash(msg)
	require.NoError(t, err)
	assert.False(key.Verify(msg, Sigs{sig3}))
	assert.False(key2.Verify(msg, Sigs{sig3}))
	assert.True(key.Verify(msg, Sigs{sig3, sig2, sig}))
}

func TestKeys(t *testing.T) {
	assert := assert.New(t)

	var keys Keys

	assert.False(keys.Has(Signer1.Key()))
	assert.False(keys.HasAddress(Signer1.Key().Address()))
	assert.True(keys.FindKeyOrAddr(Signer1.Key().Address()).Equal(Signer1.Key()), "Secp256k1 key")
	assert.False(keys.FindKeyOrAddr(Signer3.Key().Address()).Equal(Signer3.Key()), "Ed25519 key")
	assert.Nil(keys.Valid())
	assert.Nil(keys)
	assert.Nil(keys.Clone())
	assert.Equal(Keys{}, Keys{}.Clone())

	msg := encoding.Sum256([]byte("hello"))
	sig, err := Signer1.SignHash(msg)
	require.NoError(t, err)
	assert.False(keys.Verify(msg, Sigs{sig}, 0))
	assert.False(keys.VerifyPlus(msg, Sigs{sig}, 0))
	assert.False(keys.Verify(msg, Sigs{sig}, 1))
	assert.False(keys.VerifyPlus(msg, Sigs{sig}, 1))

	keys = Keys{Signer1.Key(), Signer2.Key(), Signer3.Key()}
	assert.True(keys.Has(Signer1.Key()))
	assert.True(keys.Has(Signer3.Key()))
	assert.False(keys.Has(Signer4.Key()))

	assert.True(keys.HasAddress(Signer1.Key().Address()))
	assert.True(keys.HasAddress(Signer3.Key().Address()))
	assert.False(keys.HasAddress(Signer4.Key().Address()))

	assert.True(keys.FindKeyOrAddr(Signer1.Key().Address()).Equal(Signer1.Key()), "Secp256k1 key")
	assert.True(keys.FindKeyOrAddr(Signer3.Key().Address()).Equal(Signer3.Key()), "Ed25519 key")
	assert.False(keys.FindKeyOrAddr(Signer4.Key().Address()).Equal(Signer4.Key()), "Ed25519 key")

	invalidKeys1 := Keys{Signer1.Key(), Signer2.Key(), Signer1.Key()}
	invalidKeys2 := Keys{Signer1.Key(), Signer2.Key(), Key{}}

	assert.Nil(keys.Valid())
	assert.ErrorContains(invalidKeys1.Valid(), "duplicate key jbl8fOziScK5i9wCJsxMKle_UvwKxwPH")
	assert.ErrorContains(invalidKeys2.Valid(), "empty key")

	var key Key
	assert.ErrorContains(Keys{Signer1.Key(), key}.Valid(), "empty key")
	assert.ErrorContains(Keys{Signer1.Key(), Key{1, 2, 3, 4}}.Valid(), "invalid key AQIDBJZtvcs")

	assert.Equal(keys, keys.Clone())
	assert.Equal(invalidKeys1, invalidKeys1.Clone())
	assert.Equal(invalidKeys2, invalidKeys2.Clone())

	msg = encoding.Sum256([]byte("LDC Labs"))
	sig1, err := Signer1.SignHash(msg)
	require.NoError(t, err)
	sig2, err := Signer2.SignHash(msg)
	require.NoError(t, err)
	sig3, err := Signer4.SignHash(msg)
	require.NoError(t, err)

	assert.False(keys.Verify(msg, Sigs{sig1}, 0))
	assert.True(keys.VerifyPlus(msg, Sigs{sig1}, 0))

	assert.True(keys.Verify(msg, Sigs{sig1}, 1))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1}, 1))
	assert.True(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig3}, 1))

	assert.False(keys.Verify(msg, Sigs{sig1}, 2))
	assert.True(keys.Verify(msg, Sigs{sig1, sig2}, 2))
	assert.False(keys.Verify(msg, Sigs{sig1, sig2}, 3))
	assert.False(keys.Verify(msg, Sigs{sig1, sig2, sig3}, 3))

	assert.False(keys.VerifyPlus(msg, Sigs{sig1}, 2))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2}, 2))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig3}, 2))

	keys = append(keys, Signer4.Key())
	assert.True(keys.Verify(msg, Sigs{sig1, sig2, sig3}, 3))
	assert.True(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig3}, 2))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig3}, 3))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig3, Sig{}}, 3))

	// duplicate sig
	assert.False(keys.Verify(msg, Sigs{sig1, sig2, sig2}, 3))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig2}, 2))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig2}, 3))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig2, Sig{}}, 3))

	// duplicate key and sig
	keys[3] = Signer2.Key()
	assert.False(keys.Verify(msg, Sigs{sig1, sig2, sig2}, 3))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig2}, 2))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig2}, 3))
	assert.False(keys.VerifyPlus(msg, Sigs{sig1, sig2, sig2, Sig{}}, 3))
}

func TestKeyInStruct(t *testing.T) {
	assert := assert.New(t)

	type T1 struct {
		Key *Key `cbor:"k,omitempty"`
	}

	t1 := &T1{Key: nil}
	d, err := encoding.MarshalCBOR(t1)
	require.NoError(t, err)
	assert.Equal("a0", fmt.Sprintf("%x", d))

	var t1b T1
	assert.NoError(encoding.UnmarshalCBOR(d, &t1b))

	var key Key
	t1 = &T1{Key: &key}
	d, err = encoding.MarshalCBOR(t1)
	require.NoError(t, err)
	assert.Equal("a1616bf6", fmt.Sprintf("%x", d))
	assert.NoError(encoding.UnmarshalCBOR(d, &t1b))
	assert.True(t1b.Key == nil)
	assert.Equal("a0", fmt.Sprintf("%x", encoding.MustMarshalCBOR(t1b))) // fix?

	t1 = &T1{Key: &Key{}}
	d, err = encoding.MarshalCBOR(t1)
	require.NoError(t, err)
	assert.Equal("a1616b40", fmt.Sprintf("%x", d))
	assert.NoError(encoding.UnmarshalCBOR(d, &t1b))
	assert.True(t1b.Key != nil)
	assert.Equal("a1616b40", fmt.Sprintf("%x", encoding.MustMarshalCBOR(t1b)))
	assert.NoError(t1b.Key.ValidOrEmpty())

	type T2 struct {
		Key Key `cbor:"k,omitempty"`
	}

	t2 := &T2{}
	d, err = encoding.MarshalCBOR(t2)
	require.NoError(t, err)
	assert.Equal("a1616bf6", fmt.Sprintf("%x", d))

	var t2b T2
	assert.NoError(encoding.UnmarshalCBOR(d, &t2b))
	assert.Nil(t2b.Key)

	t2 = &T2{Key: key}
	d, err = encoding.MarshalCBOR(t2)
	require.NoError(t, err)
	assert.Equal("a1616bf6", fmt.Sprintf("%x", d))
	assert.NoError(encoding.UnmarshalCBOR(d, &t2b))
	assert.True(t2b.Key == nil)
	assert.Equal("a1616bf6", fmt.Sprintf("%x", encoding.MustMarshalCBOR(t2b)))
	assert.ErrorContains(t2b.Key.ValidOrEmpty(), "nil key")

	t2 = &T2{Key: Key{}}
	d, err = encoding.MarshalCBOR(t2)
	require.NoError(t, err)
	assert.Equal("a1616b40", fmt.Sprintf("%x", d))
	assert.NoError(encoding.UnmarshalCBOR(d, &t2b))
	assert.True(t2b.Key != nil)
	assert.Equal("a1616b40", fmt.Sprintf("%x", encoding.MustMarshalCBOR(t2b)))
	assert.NoError(t2b.Key.ValidOrEmpty())
}
