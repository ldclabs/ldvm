// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDN(t *testing.T) {
	assert := assert.New(t)

	dn, err := NewDN("")
	assert.ErrorContains(err, `NewDN(""): invalid utf8 name`)
	assert.Nil(dn)

	dn, err = NewDN(".")
	assert.ErrorContains(err, `NewDN("."): ToASCII error, idna: invalid label "."`)
	assert.Nil(dn)

	dn, err = NewDN(".com")
	assert.ErrorContains(err, `NewDN(".com"): ToASCII error, idna: invalid label ".com"`)
	assert.Nil(dn)

	dn, err = NewDN("abc.com")
	assert.ErrorContains(err, `NewDN("abc.com"): invalid domain name, no trailing dot`)
	assert.Nil(dn)

	dn, err = NewDN("abc..com")
	assert.ErrorContains(err, `NewDN("abc..com"): ToASCII error, idna: invalid label "abc..com"`)
	assert.Nil(dn)

	dn, err = NewDN("abc_com")
	assert.ErrorContains(err, `NewDN("abc_com"): ToASCII error, idna: disallowed rune`)
	assert.Nil(dn)

	dn, err = NewDN("abc com")
	assert.ErrorContains(err, `NewDN("abc com"): ToASCII error, idna: disallowed rune`)
	assert.Nil(dn)

	dn, err = NewDN("com.")
	assert.NoError(err)
	assert.True(dn.IsDomain())
	assert.Equal("com.", dn.ASCII())
	assert.Equal("com.", dn.String())

	dn, err = NewDN("abc.com.")
	assert.NoError(err)
	assert.True(dn.IsDomain())
	assert.Equal("abc.com.", dn.ASCII())
	assert.Equal("abc.com.", dn.String())

	dn, err = NewDN("公信.com.")
	assert.NoError(err)
	assert.True(dn.IsDomain())
	assert.Equal("xn--vuq70b.com.", dn.ASCII())
	assert.Equal("公信.com.", dn.String())

	dn, err = NewDN("xn--vuq70b.公信.")
	assert.ErrorContains(err, `NewDN("xn--vuq70b.公信."): invalid domain name`)
	assert.Nil(dn)

	dn, err = NewDN(":")
	assert.ErrorContains(err, `NewDN(":"): ToASCII error, idna: invalid label ""`)
	assert.Nil(dn)

	dn, err = NewDN("did:")
	assert.ErrorContains(err, `NewDN("did:"): ToASCII error, idna: invalid label ""`)
	assert.Nil(dn)

	dn, err = NewDN(":did")
	assert.ErrorContains(err, `NewDN(":did"): ToASCII error, idna: invalid label ""`)
	assert.Nil(dn)

	dn, err = NewDN("did::ldc")
	assert.ErrorContains(err, `NewDN("did::ldc"): ToASCII error, idna: invalid label ""`)
	assert.Nil(dn)

	dn, err = NewDN("com")
	assert.NoError(err)
	assert.False(dn.IsDomain())
	assert.Equal("com", dn.ASCII())
	assert.Equal("com", dn.String())

	dn, err = NewDN("公信")
	assert.NoError(err)
	assert.False(dn.IsDomain())
	assert.Equal("xn--vuq70b", dn.ASCII())
	assert.Equal("公信", dn.String())

	dn, err = NewDN("xn--vuq70b")
	assert.NoError(err)
	assert.False(dn.IsDomain())
	assert.Equal("xn--vuq70b", dn.ASCII())
	assert.Equal("公信", dn.String())

	dn, err = NewDN("did:公信")
	assert.NoError(err)
	assert.False(dn.IsDomain())
	assert.Equal("did:xn--vuq70b", dn.ASCII())
	assert.Equal("did:公信", dn.String())

	dn, err = NewDN("公信:公信")
	assert.NoError(err)
	assert.False(dn.IsDomain())
	assert.Equal("xn--vuq70b:xn--vuq70b", dn.ASCII())
	assert.Equal("公信:公信", dn.String())

	dn, err = NewDN("xn--vuq70b:xn--vuq70b")
	assert.NoError(err)
	assert.False(dn.IsDomain())
	assert.Equal("xn--vuq70b:xn--vuq70b", dn.ASCII())
	assert.Equal("公信:公信", dn.String())

	dn, err = NewDN("公信:xn--vuq70b")
	assert.ErrorContains(err, `NewDN("公信:xn--vuq70b"): invalid decentralized name`)
	assert.Nil(dn)
}
