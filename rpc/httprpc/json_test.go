// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package httprpc

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
	"github.com/ldclabs/ldvm/util/httpcli"
	"github.com/ldclabs/ldvm/util/sync"
	"github.com/ldclabs/ldvm/util/value"
)

type jsonhandler struct{}

type jsonResult struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func (h *jsonhandler) ServeRPC(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {

	value.DoIfCtxValueValid(ctx, func(log *value.Log) {
		log.Set("rpcId", value.String(req.ID))
		log.Set("rpcMethod", value.String(req.Method))
	})

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

type httpjsonhandler struct {
	handler http.Handler
	snap    bool
	r       sync.Value[*http.Request]
	wh      sync.Value[http.Header]
	ctx     sync.Value[context.Context]
}

func (h *httpjsonhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := value.CtxWith(r.Context(), &value.Log{Value: value.NewMap(16)})
	r = r.WithContext(ctx)

	if h.snap {
		h.r.Store(r)
		h.wh.Store(w.Header())
		h.ctx.Store(ctx)
	}

	h.handler.ServeHTTP(w, r)
}

func TestJSONRPC(t *testing.T) {
	assert := assert.New(t)

	ch := &jsonhandler{}
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

		assert.Equal("TestMethod", re.Method)
		assert.Equal(mustMarshalJSON(params), []byte(re.Params))

		assert.NotNil(hh.r.Load())
		assert.Equal(jsonrpc.MIMEApplicationJSON, hh.r.MustLoad().Header.Get("Accept"))
		assert.Equal(jsonrpc.MIMEApplicationJSONCharsetUTF8, hh.r.MustLoad().Header.Get("Content-Type"))
		assert.Equal("zstd,gzip", hh.r.MustLoad().Header.Get("Accept-Encoding"))
		assert.Equal("TestJSONRPC", hh.r.MustLoad().Header.Get("User-Agent"))

		assert.Equal("HTTP/2.0", hh.r.MustLoad().Proto)
		assert.Equal(res.ID, hh.r.MustLoad().Header.Get("X-Request-ID"))

		assert.NotNil(hh.wh.Load())
		assert.Equal(jsonrpc.MIMEApplicationJSONCharsetUTF8, hh.wh.MustLoad().Get("Content-Type"))
		assert.Equal("nosniff", hh.wh.MustLoad().Get("x-Content-Type-Options"))
		assert.Equal("zstd", hh.wh.MustLoad().Get("Content-Encoding"))
		assert.Equal("4186", hh.wh.MustLoad().Get("X-Content-Length"))
		assert.Equal(res.ID, hh.wh.MustLoad().Get("X-Request-ID"))

		log := value.CtxValue[value.Log](hh.ctx.MustLoad()).Map()
		assert.Equal([]string{"elapsed", "method", "proto", "remoteAddr", "requestBytes", "requestUri", "responseBytes", "rpcId", "rpcMethod", "start", "status", "user-agent", "x-request-id"}, log.Keys())

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
		assert.Equal(
			`{"code":-32000,"message":"context canceled","errors":[],"data":null}`,
			res.Error.Error())

		res = cli.Do(context.Background(), req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		assert.Equal(
			`{"code":-32601,"message":"method \"\" not found","errors":[],"data":null}`,
			res.Error.Error())

		req = &jsonrpc.Request{ID: "abcd", Method: "ErrorMethod"}
		res = cli.Do(context.Background(), req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		assert.Equal(
			`{"code":-32601,"message":"method \"ErrorMethod\" not found","errors":[],"data":null}`,
			res.Error.Error())
		log := value.CtxValue[value.Log](hh.ctx.MustLoad()).Map()
		assert.Equal(res.Error.Error(), log["responseError"].String())

		req = &jsonrpc.Request{ID: "abcd", Method: "Get"}
		res = cli.Do(context.Background(), req)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
		assert.Equal("abcd", res.ID)
		assert.Equal(
			`{"code":-32602,"message":"invalid parameter(s), no params","errors":[],"data":null}`,
			res.Error.Error())
		log = value.CtxValue[value.Log](hh.ctx.MustLoad()).Map()
		assert.Equal(res.Error.Error(), log["responseError"].String())
	})
}

func TestJSONRPCChaos(t *testing.T) {
	ch := &jsonhandler{}
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
