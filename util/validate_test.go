// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidName(t *testing.T) {
	assert := assert.New(t)

	assert.False(ValidName(""))

	assert.True(ValidName("a"))
	assert.True(ValidName("a正"))
	assert.True(ValidName("a🀄️"))
	assert.True(ValidName("Hello world!"))

	assert.False(ValidName(" a"))
	assert.False(ValidName("a "))
	assert.False(ValidName("a\na"))
	assert.False(ValidName("a\bb"))
	assert.False(ValidName("a  a"))
	assert.False(ValidName("a\ta"))
	assert.False(ValidName("Hello‍ world!"))
}

func TestValidLink(t *testing.T) {
	assert := assert.New(t)

	assert.True(ValidLink(""))
	assert.True(ValidLink("0x0000000000000000000000000000000000000000"))
	assert.True(ValidLink("LM1111111111111111111Ax1asG"))
	assert.True(ValidLink("mail:to"))
	assert.True(ValidLink("https://hello.com/abc"))

	assert.False(ValidLink("正"))
	assert.False(ValidLink("mail:‍to"))
	assert.False(ValidLink("https://hello.com/ab%c"))
}

func TestValidMessage(t *testing.T) {
	assert := assert.New(t)

	assert.True(ValidMessage(""))
	assert.True(ValidMessage("Hello, world!"))
	assert.True(ValidMessage("你好👋"))

	assert.False(ValidMessage("‍"))              // Zero Width Joiner, https://emojipedia.org/zero-width-joiner/
	assert.False(ValidMessage("Hello‍, world!")) // with Zero Width Joiner
	assert.False(ValidMessage(" Hello, world!"))
	assert.False(ValidMessage("\nHello, world!"))
}
