// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"fmt"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

// https://emojipedia.org/zero-width-joiner/
func TestValidName(t *testing.T) {
	assert := assert.New(t)

	assert.True(ValidName("a"))
	assert.True(ValidName("a正"))
	assert.True(ValidName("a 正"))
	assert.True(ValidName("🀄正"))
	assert.True(ValidName("Hello world!❤️‍🔥🧑‍🤝‍🧑"))

	assert.False(ValidName(""))
	assert.False(ValidName("‍"))
	assert.False(ValidName(" a"))
	assert.False(ValidName("a "))
	assert.False(ValidName("a‍"))
	assert.False(ValidName("a  正"))
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

	assert.False(ValidMessage("‍"))              // Zero Width Joiner
	assert.False(ValidMessage("Hello‍, world!")) // with Zero Width Joiner
	assert.False(ValidMessage(" Hello, world!"))
	assert.False(ValidMessage("\nHello, world!"))
}

func printRune(r rune) {
	fmt.Println(unicode.IsControl(r), "IsControl")
	fmt.Println(unicode.IsDigit(r), "IsDigit")
	fmt.Println(unicode.IsGraphic(r), "IsGraphic")
	fmt.Println(unicode.IsLetter(r), "IsLetter")
	fmt.Println(unicode.IsLower(r), "IsLower")
	fmt.Println(unicode.IsMark(r), "IsMark")
	fmt.Println(unicode.IsNumber(r), "IsNumber")
	fmt.Println(unicode.IsPrint(r), "IsPrint")
	fmt.Println(unicode.IsPunct(r), "IsPunct")
	fmt.Println(unicode.IsSpace(r), "IsSpace")
	fmt.Println(unicode.IsSymbol(r), "IsSymbol")
	fmt.Println(unicode.IsTitle(r), "IsTitle")
	fmt.Println(unicode.IsUpper(r), "IsUpper")
}
