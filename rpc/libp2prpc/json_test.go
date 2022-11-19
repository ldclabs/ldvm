// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/rpc/protocol/jsonrpc"
)

type jsonhandler struct {
	snap bool
	err  atomic.Value // *jsonrpc.Error
}

type jsonresult struct {
	Method string          `cbor:"method"`
	Params json.RawMessage `cbor:"params"`
}

func (h *jsonhandler) ServeRPC(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	switch {
	case req.Method == "ErrorMethod":
		return req.InvalidMethod()

	case req.Params == nil:
		return req.InvalidParams("no params")

	default:
		return req.Result(&jsonresult{
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

func TestJSONRPC(t *testing.T) {
	paddrs := [][]string{
		{"tcp/23581", "tcp/23582"},
		{"udp/23581/quic", "udp/23582/quic"},
		{"tcp/23583/ws", "tcp/23584/ws"},
	}

	for _, pa := range paddrs {
		t.Run(pa[0], func(t *testing.T) {
			ch := &jsonhandler{snap: true}
			ha1, err := makeBasicHost(pa[0])
			require.NoError(t, err)

			ha2, err := makeBasicHost(pa[1])
			require.NoError(t, err)
			ha2.Peerstore().AddAddrs(ha1.ID(), ha1.Addrs(), peerstore.PermanentAddrTTL)

			_ = NewJSONService(ha1, ch, nil)
			cli := NewJSONClient(ha2, ha1.ID(), &JSONClientOptions{Compress: false})
			cli2 := NewJSONClient(ha2, ha1.ID(), &JSONClientOptions{Compress: true})

			defer ha1.Close()
			defer ha2.Close()

			t.Run("should work", func(t *testing.T) {
				assert := assert.New(t)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				params := strings.Repeat("test", 1024)
				re := &jsonresult{}
				res := cli.Request(ctx, "TestMethod", params, re)
				require.Nil(t, res.Error)

				assert.Nil(ch.err.Load())
				assert.Equal("TestMethod", re.Method)
				assert.Equal(mustMarshalJSON(params), []byte(re.Params))

				re = &jsonresult{}
				res = cli2.Request(ctx, "TestMethod", params, re)
				require.Nil(t, res.Error)

				assert.Nil(ch.err.Load())
				assert.Equal("TestMethod", re.Method)
				assert.Equal(mustMarshalJSON(params), []byte(re.Params))

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
					re := &jsonresult{}
					res := cli.Request(ctx, "Echo", params, re)
					require.Nil(t, res.Error)

					assert.Equal("Echo", re.Method)
					assert.Equal(mustMarshalJSON(params), []byte(re.Params))

					re = &jsonresult{}
					res = cli2.Request(ctx, "Echo", params, re)
					require.Nil(t, res.Error)

					assert.Equal("Echo", re.Method)
					assert.Equal(mustMarshalJSON(params), []byte(re.Params))
				}
			})

			t.Run("error case", func(t *testing.T) {
				assert := assert.New(t)

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
		})
	}
}

func TestJSONRPCChaos(t *testing.T) {
	paddrs := [][]string{
		{"tcp/23585", "tcp/23586"},
		{"udp/23585/quic", "udp/23586/quic"},
		{"tcp/23587/ws", "tcp/23588/ws"},
	}

	for _, pa := range paddrs {
		t.Run(pa[0], func(t *testing.T) {
			assert := assert.New(t)

			ch := &jsonhandler{snap: false}
			ha1, err := makeBasicHost(pa[0])
			require.NoError(t, err)
			ha1.Network().ResourceManager().Close()

			ha2, err := makeBasicHost(pa[1])
			require.NoError(t, err)
			ha2.Peerstore().AddAddrs(ha1.ID(), ha1.Addrs(), peerstore.PermanentAddrTTL)

			_ = NewJSONService(ha1, ch, nil)
			cli := NewJSONClient(ha2, ha1.ID(), nil)

			defer ha1.Close()
			defer ha2.Close()

			wg := &sync.WaitGroup{}
			total := 10

			for i := 0; i < total; i++ {
				wg.Add(1)
				go func(x int) {
					defer wg.Done()
					re := &jsonresult{}
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
