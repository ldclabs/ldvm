// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/httpcli"
)

type cborhandler struct {
	snap bool
	err  *cborrpc.Error
}

type result struct {
	Method string          `cbor:"method"`
	Params cbor.RawMessage `cbor:"params"`
}

func (h *cborhandler) ServeRPC(ctx context.Context, req *cborrpc.Request) *cborrpc.Response {
	switch {
	case req.Method == "ErrorMethod":
		return req.InvalidMethod()

	case req.Params == nil:
		return req.InvalidParams("no params")

	default:
		return req.Result(&result{
			Method: req.Method,
			Params: req.Params,
		})
	}
}

func (h *cborhandler) OnError(ctx context.Context, err *cborrpc.Error) {
	if h.snap {
		h.err = err
	}
}

type httphandler struct {
	handler http.Handler
	snap    bool
	r       *http.Request
	wh      http.Header
}

func (h *httphandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.snap {
		h.wh = w.Header()
		h.r = r
	}

	h.handler.ServeHTTP(w, r)
}

func TestCBORRPC(t *testing.T) {
	assert := assert.New(t)

	ch := &cborhandler{snap: true}
	hh := &httphandler{handler: NewCBORService(ch), snap: true}
	server := httpcli.NewHTTPServer(hh)
	defer server.Close()

	rt, err := httpcli.NewRoundTripper(&httpcli.TransportOptions{})
	require.NoError(t, err)

	url := "h2c://" + server.Addr().String()
	header := http.Header{}
	header.Set("user-agent", "TestCBORRPC")
	cli := NewCBORClient(url, rt, header)

	t.Run("Request should work", func(t *testing.T) {
		re := &result{}
		res := cli.Request(context.Background(), "TestMethod", 1234, re)
		require.Nil(t, res.Error)

		assert.Nil(ch.err)
		assert.Equal("TestMethod", re.Method)
		assert.Equal(encoding.MustMarshalCBOR(1234), []byte(re.Params))

		assert.NotNil(hh.r)
		assert.Equal(cborrpc.MIMEApplicationCBOR, hh.r.Header.Get("Accept"))
		assert.Equal(cborrpc.MIMEApplicationCBOR, hh.r.Header.Get("Content-Type"))
		assert.Equal("gzip", hh.r.Header.Get("Accept-Encoding"))
		assert.Equal("TestCBORRPC", hh.r.Header.Get("User-Agent"))

		assert.Equal("HTTP/2.0", hh.r.Proto)
		assert.Equal(res.ID, hh.r.Header.Get("X-Request-ID"))

		assert.NotNil(hh.wh)
		assert.Equal(cborrpc.MIMEApplicationCBOR, hh.wh.Get("Content-Type"))
		assert.Equal("nosniff", hh.wh.Get("x-Content-Type-Options"))
		assert.Equal(res.ID, hh.wh.Get("X-Request-ID"))
	})

	t.Run("error case", func(t *testing.T) {
		req := &cborrpc.Request{ID: "abcd"}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		res := cli.Do(ctx, req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		assert.ErrorContains(res.Error, `{"code":-32000,"message":"context canceled"}`)

		res = cli.Do(context.Background(), req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		assert.ErrorContains(res.Error, `{"code":-32601,"message":"method \"\" not found"}`)

		req = &cborrpc.Request{ID: "abcd", Method: "ErrorMethod"}
		res = cli.Do(context.Background(), req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		assert.Equal(ch.err.Error(), res.Error.Error())
		assert.Equal(`{"code":-32601,"message":"method \"ErrorMethod\" not found"}`, res.Error.Error())

		req = &cborrpc.Request{ID: "abcd", Method: "Get"}
		res = cli.Do(context.Background(), req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		assert.Equal(ch.err.Error(), res.Error.Error())
		assert.Equal(`{"code":-32602,"message":"invalid parameter(s), no params"}`, res.Error.Error())
	})
}

func TestCBORRPCChaos(t *testing.T) {
	assert := assert.New(t)

	ch := &cborhandler{snap: false}
	hh := &httphandler{handler: NewCBORService(ch), snap: false}
	server := httpcli.NewHTTPServer(hh)
	defer server.Close()

	rt, err := httpcli.NewRoundTripper(&httpcli.TransportOptions{})
	require.NoError(t, err)

	url := "h2c://" + server.Addr().String()
	header := http.Header{}
	cli := NewCBORClient(url, rt, header)

	wg := &sync.WaitGroup{}
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			re := &result{}
			res := cli.Request(context.Background(), "TestMethod", x, re)
			require.Nil(t, res.Error)
			assert.Equal("TestMethod", re.Method)
			assert.Equal(encoding.MustMarshalCBOR(x), []byte(re.Params))
			data := encoding.MustMarshalCBOR(re)
			xid := res.ID
			assert.Equal(data, []byte(res.Result))

			time.Sleep(time.Duration(rand.Int63n(int64(x%9999) + 1)))
			req := &cborrpc.Request{Method: "TestMethod", Params: encoding.MustMarshalCBOR(x)}
			res = cli.Do(context.Background(), req)
			require.Nil(t, res.Error)

			assert.NotEqual("", req.ID)
			assert.NotEqual(xid, res.ID)
			assert.Equal(req.ID, res.ID)
			assert.Equal(data, []byte(res.Result))
		}(i)
	}

	wg.Wait()
}
