// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package erring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixErr(t *testing.T) {
	assert := assert.New(t)

	mocker := func(data []byte, err error) ([]byte, error) {
		return data, err
	}

	var err error
	errp := ErrPrefix("")
	assert.Nil(errp.ErrorIf(err))
	assert.EqualError(errp.ErrorIf(errors.New("invalid name")), "invalid name")
	assert.EqualError(errp.Errorf("invalid name %s", "-_*"), "invalid name -_*")

	data, er := errp.ErrorMap(mocker([]byte{0}, errors.New("invalid data")))
	assert.Nil(data)
	assert.EqualError(er, "invalid data")
	data, er = errp.ErrorMap(mocker([]byte{0}, err))
	assert.Nil(er)
	assert.Equal([]byte{0}, data)

	errp = ErrPrefix("ErrPrefix: ")
	assert.Nil(errp.ErrorIf(err))
	assert.EqualError(errp.ErrorIf(errors.New("invalid name")), "ErrPrefix: invalid name")
	assert.EqualError(errp.Errorf("invalid name %s", "-_*"), "ErrPrefix: invalid name -_*")

	data, er = errp.ErrorMap(mocker([]byte{0}, errors.New("invalid data")))
	assert.Nil(data)
	assert.EqualError(er, "ErrPrefix: invalid data")
	data, er = errp.ErrorMap(mocker([]byte{0}, err))
	assert.Nil(er)
	assert.Equal([]byte{0}, data)
}
