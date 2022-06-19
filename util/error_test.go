// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixErr(t *testing.T) {
	assert := assert.New(t)

	var err error
	errp := ErrPrefix("")
	assert.Nil(errp.ErrorIf(err))
	assert.EqualError(errp.ErrorIf(errors.New("invalid name")), "invalid name")
	assert.EqualError(errp.Errorf("invalid name %s", "-_*"), "invalid name -_*")

	errp = ErrPrefix("ErrPrefix error: ")
	assert.Nil(errp.ErrorIf(err))
	assert.EqualError(errp.ErrorIf(errors.New("invalid name")), "ErrPrefix error: invalid name")
	assert.EqualError(errp.Errorf("invalid name %s", "-_*"), "ErrPrefix error: invalid name -_*")
}
