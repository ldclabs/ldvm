// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
	"github.com/ldclabs/ldvm/util/httpcli"
)

type jsonhandler struct {
	snap bool
	err  atomic.Value // *jsonrpc.Error
}

type jsonResult struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func (h *jsonhandler) ServeRPC(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	switch {
	case req.Method == "ErrorMethod":
		return req.InvalidMethod()

	case req.Params == nil:
		return req.InvalidParams("no params")

	default:
		return req.Result(&jsonResult{
			Method: req.Method,
			Params: req.Params,
		})
	}
}

func (h *jsonhandler) OnError(ctx context.Context, err *jsonrpc.Error) {
	if h.snap {
		h.err.Store(err)
	}
}

type httpjsonhandler struct {
	handler http.Handler
	snap    bool
	r       atomic.Value // *http.Request
	wh      atomic.Value // http.Header
}

func (h *httpjsonhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.snap {
		h.wh.Store(w.Header())
		h.r.Store(r)
	}

	h.handler.ServeHTTP(w, r)
}

func TestJSONRPC(t *testing.T) {
	assert := assert.New(t)

	ch := &jsonhandler{snap: true}
	hh := &httpjsonhandler{handler: NewJSONService(ch, nil), snap: true}
	server := httpcli.NewHTTPServer(hh)
	defer server.Close()

	url := "h2c://" + server.Addr().String()
	header := http.Header{}
	header.Set("user-agent", "TestJSONRPC")
	cli := NewJSONClient(url, &JSONClientOptions{
		RoundTripper: httpcli.DefaultTransport,
		Header:       header,
	})

	t.Run("should work", func(t *testing.T) {
		re := &jsonResult{}

		params := strings.Repeat("test", 1024)
		res := cli.Request(context.Background(), "TestMethod", params, re)
		require.Nil(t, res.Error)

		assert.Nil(ch.err.Load())
		assert.Equal("TestMethod", re.Method)
		assert.Equal(mustMarshalJSON(params), []byte(re.Params))

		assert.NotNil(hh.r.Load())
		assert.Equal(jsonrpc.MIMEApplicationJSON, hh.r.Load().(*http.Request).Header.Get("Accept"))
		assert.Equal(jsonrpc.MIMEApplicationJSONCharsetUTF8, hh.r.Load().(*http.Request).Header.Get("Content-Type"))
		assert.Equal("zstd,gzip", hh.r.Load().(*http.Request).Header.Get("Accept-Encoding"))
		assert.Equal("TestJSONRPC", hh.r.Load().(*http.Request).Header.Get("User-Agent"))

		assert.Equal("HTTP/2.0", hh.r.Load().(*http.Request).Proto)
		assert.Equal(res.ID, hh.r.Load().(*http.Request).Header.Get("X-Request-ID"))

		assert.NotNil(hh.wh.Load())
		assert.Equal(jsonrpc.MIMEApplicationJSONCharsetUTF8, hh.wh.Load().(http.Header).Get("Content-Type"))
		assert.Equal("nosniff", hh.wh.Load().(http.Header).Get("x-Content-Type-Options"))
		assert.Equal("zstd", hh.wh.Load().(http.Header).Get("Content-Encoding"))
		assert.Equal("4186", hh.wh.Load().(http.Header).Get("X-Content-Length"))
		assert.Equal(res.ID, hh.wh.Load().(http.Header).Get("X-Request-ID"))

		cases := []interface{}{
			0,
			123,
			-123,
			"",
			"abc",
			true,
			false,
			[]byte{255, 254, 253, 0},
			[]ids.ID20{ids.EmptyID20, {1}, {2}},
			[]string{"a", "b", "c"},
			map[int]string{0: "a", 1: "b"},
		}

		for _, params := range cases {
			re := &jsonResult{}
			res := cli.Request(context.Background(), "Echo", params, re)
			require.Nil(t, res.Error)

			assert.Equal("Echo", re.Method)
			assert.Equal(mustMarshalJSON(params), []byte(re.Params))
		}
	})

	t.Run("error case", func(t *testing.T) {
		req := &jsonrpc.Request{ID: "abcd"}
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

		req = &jsonrpc.Request{ID: "abcd", Method: "ErrorMethod"}
		res = cli.Do(context.Background(), req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		require.NotNil(t, ch.err.Load())
		assert.Equal(ch.err.Load().(error).Error(), res.Error.Error())
		fmt.Println(res.Error.Error())
		assert.Equal(`{"code":-32601,"message":"method \"ErrorMethod\" not found"}`, res.Error.Error())

		req = &jsonrpc.Request{ID: "abcd", Method: "Get"}
		res = cli.Do(context.Background(), req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		require.NotNil(t, ch.err.Load())
		assert.Equal(ch.err.Load().(error).Error(), res.Error.Error())
		assert.Equal(`{"code":-32602,"message":"invalid parameter(s), no params"}`, res.Error.Error())
	})
}

func TestJSONRPCChaos(t *testing.T) {
	ch := &jsonhandler{snap: false}
	hh := &httpjsonhandler{handler: NewJSONService(ch, nil), snap: false}
	server := httpcli.NewHTTPServer(hh)
	defer server.Close()

	schemes := []string{"http", "h2c"}

	for _, s := range schemes {
		t.Run(s, func(t *testing.T) {
			assert := assert.New(t)
			url := s + "://" + server.Addr().String()
			cli := NewJSONClient(url, nil)

			wg := &sync.WaitGroup{}
			total := 100
			if s == "h2c" {
				total = 10000
			}

			for i := 0; i < total; i++ {
				wg.Add(1)
				go func(x int) {
					defer wg.Done()
					re := &jsonResult{}
					res := cli.Request(context.Background(), "TestMethod", x, re)
					require.Nil(t, res.Error)
					assert.Equal("TestMethod", re.Method)
					assert.Equal(mustMarshalJSON(x), []byte(re.Params))
					data := mustMarshalJSON(re)
					xid := res.ID
					assert.Equal(data, []byte(res.Result))

					time.Sleep(time.Duration(rand.Int63n(int64(x%999) + 1)))
					req := &jsonrpc.Request{Method: "TestMethod", Params: mustMarshalJSON(x)}
					res = cli.Do(context.Background(), req)
					require.Nil(t, res.Error)

					assert.NotEqual("", req.ID)
					assert.NotEqual(xid, res.ID)
					assert.Equal(req.ID, res.ID)
					assert.Equal(data, []byte(res.Result))
				}(i)
			}

			wg.Wait()
		})
	}
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
