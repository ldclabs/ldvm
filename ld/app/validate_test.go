// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidDomainName(t *testing.T) {
	assert.False(t, ValidDomainName(""))

	assert.True(t, ValidDomainName("aa"))
	assert.True(t, ValidDomainName("Aa"))
	assert.True(t, ValidDomainName("aa"))
	assert.True(t, ValidDomainName("a-a"))
	assert.True(t, ValidDomainName("a.a"))
	assert.True(t, ValidDomainName("a.com"))

	assert.False(t, ValidDomainName("a/a"))
	assert.False(t, ValidDomainName("a_a"))
	assert.False(t, ValidDomainName("a正"))
	assert.False(t, ValidDomainName("a a"))
	assert.False(t, ValidDomainName("a."))
	assert.False(t, ValidDomainName(".a"))
	assert.False(t, ValidDomainName("a-"))
	assert.False(t, ValidDomainName("-a"))
	assert.False(t, ValidDomainName("a..a"))
	assert.False(t, ValidDomainName("a-.a"))
	assert.False(t, ValidDomainName("a.com."))
}

func TestValidName(t *testing.T) {
	assert.False(t, ValidName(""))

	assert.True(t, ValidName("a"))
	assert.True(t, ValidName("a正"))
	assert.True(t, ValidName("a🀄️"))
	assert.True(t, ValidName("Hello world!"))

	assert.False(t, ValidName(" a"))
	assert.False(t, ValidName("a "))
	assert.False(t, ValidName("a\na"))
	assert.False(t, ValidName("a\bb"))
	assert.False(t, ValidName("a  a"))
	assert.False(t, ValidName("a\ta"))
}

func TestValidLink(t *testing.T) {
	assert.True(t, ValidLink(""))
	assert.True(t, ValidLink("0x0000000000000000000000000000000000000000"))
	assert.True(t, ValidLink("LM1111111111111111111Ax1asG"))
	assert.True(t, ValidLink("mail:to"))
	assert.True(t, ValidLink("https://hello.com/abc"))

	assert.False(t, ValidLink("正"))
	assert.False(t, ValidLink("https://hello.com/ab%c"))
}

func TestValidMID(t *testing.T) {
	assert.True(t, ValidMID(""))
	assert.True(t, ValidMID("LM1111111111111111111Ax1asG"))

	assert.False(t, ValidMID("LD1111111111111111111Ax1asG"))
	assert.False(t, ValidMID("0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF"))
	assert.False(t, ValidMID("FFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF"))
}
