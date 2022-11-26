// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httpcli

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(r.Proto))
}

func TestHTTPCli(t *testing.T) {
	assert := assert.New(t)

	server := NewHTTPServer(&handler{})
	defer server.Close()

	rt, err := NewRoundTripper(&TransportOptions{})
	require.NoError(t, err)

	cli := NewClient(rt)
	url := "http://" + server.Addr().String()

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	require.NoError(t, err)

	data, err := cli.DoAndRead(req)
	require.NoError(t, err)
	assert.Equal("HTTP/1.1", string(data))

	url = "h2c://" + server.Addr().String()
	req, err = http.NewRequestWithContext(context.Background(), "GET", url, nil)
	require.NoError(t, err)

	data, err = cli.DoAndRead(req)
	require.NoError(t, err)
	assert.Equal("HTTP/2.0", string(data))
}
