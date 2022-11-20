// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package compress

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompress(t *testing.T) {
	assert := assert.New(t)

	var w, r bytes.Buffer

	data := []byte(strings.Repeat("hello", 1000))

	zw := NewZstdWriter(&w)
	n, err := zw.Write(data)
	require.NoError(t, err)
	assert.Equal(len(data), n)
	zw.Reset()
	assert.True(len(data) > w.Len())

	zr := NewZstdReader(&w)
	_, err = r.ReadFrom(zr)
	require.NoError(t, err)
	zr.Reset()

	assert.Equal(data, r.Bytes())

	w.Reset()
	r.Reset()
	data = []byte(strings.Repeat("world", 1000))
	zw = NewZstdWriter(&w)
	n, err = zw.Write(data)
	require.NoError(t, err)
	assert.Equal(len(data), n)
	require.NoError(t, zw.Close())
	assert.True(len(data) > w.Len())

	zr = NewZstdReader(&w)
	_, err = r.ReadFrom(zr)
	require.NoError(t, err)
	require.NoError(t, zr.Close())

	assert.Equal(data, r.Bytes())
}
