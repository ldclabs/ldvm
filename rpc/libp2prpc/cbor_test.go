// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
	lsync "github.com/ldclabs/ldvm/util/sync"
)

type cborhandler struct {
	snap bool
	err  lsync.Value[*cborrpc.Error]
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
		h.err.Store(err)
	}
}

func TestCBORRPC(t *testing.T) {
	paddrs := [][]string{
		{"tcp/23571", "tcp/23572"},
		{"udp/23571/quic", "udp/23572/quic"},
		{"tcp/23573/ws", "tcp/23574/ws"},
	}

	for _, pa := range paddrs {
		t.Run(pa[0], func(t *testing.T) {
			ch := &cborhandler{snap: true}
			ha1, err := makeBasicHost(pa[0])
			require.NoError(t, err)

			ha2, err := makeBasicHost(pa[1])
			require.NoError(t, err)
			ha2.Peerstore().AddAddrs(ha1.ID(), ha1.Addrs(), peerstore.PermanentAddrTTL)

			_ = NewCBORService(ha1, ch, nil)
			cli := NewCBORClient(ha2, ha1.ID(), &CBORClientOptions{Compress: false})
			cli2 := NewCBORClient(ha2, ha1.ID(), &CBORClientOptions{Compress: true})

			defer ha1.Close()
			defer ha2.Close()

			t.Run("should work", func(t *testing.T) {
				assert := assert.New(t)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				params := strings.Repeat("test", 1024)
				re := &result{}
				res := cli.Request(ctx, "TestMethod", params, re)
				require.Nil(t, res.Error)

				assert.Nil(ch.err.Load())
				assert.Equal("TestMethod", re.Method)
				assert.Equal(encoding.MustMarshalCBOR(params), []byte(re.Params))

				re = &result{}
				res = cli2.Request(ctx, "TestMethod", params, re)
				require.Nil(t, res.Error)

				assert.Nil(ch.err.Load())
				assert.Equal("TestMethod", re.Method)
				assert.Equal(encoding.MustMarshalCBOR(params), []byte(re.Params))

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
					re := &result{}
					res := cli.Request(ctx, "Echo", params, re)
					require.Nil(t, res.Error)

					assert.Equal("Echo", re.Method)
					assert.Equal(encoding.MustMarshalCBOR(params), []byte(re.Params))

					re = &result{}
					res = cli2.Request(ctx, "Echo", params, re)
					require.Nil(t, res.Error)

					assert.Equal("Echo", re.Method)
					assert.Equal(encoding.MustMarshalCBOR(params), []byte(re.Params))
				}
			})

			t.Run("error case", func(t *testing.T) {
				assert := assert.New(t)

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
				assert.Equal(ch.err.MustLoad().Error(), res.Error.Error())
				assert.Equal(`{"code":-32601,"message":"method \"ErrorMethod\" not found"}`, res.Error.Error())

				req = &cborrpc.Request{ID: "abcd", Method: "Get"}
				res = cli.Do(context.Background(), req)
				assert.NotNil(res.Error)
				assert.Nil(res.Result)
				assert.Equal("abcd", res.ID)
				assert.Equal(ch.err.MustLoad().Error(), res.Error.Error())
				assert.Equal(`{"code":-32602,"message":"invalid parameter(s), no params"}`, res.Error.Error())
			})
		})
	}
}

func TestCBORRPCChaos(t *testing.T) {
	paddrs := [][]string{
		{"tcp/23575", "tcp/23576"},
		{"udp/23575/quic", "udp/23576/quic"},
		{"tcp/23577/ws", "tcp/23578/ws"},
	}

	for _, pa := range paddrs {
		t.Run(pa[0], func(t *testing.T) {
			assert := assert.New(t)

			ch := &cborhandler{snap: false}
			ha1, err := makeBasicHost(pa[0])
			require.NoError(t, err)
			ha1.Network().ResourceManager().Close()

			ha2, err := makeBasicHost(pa[1])
			require.NoError(t, err)
			ha2.Peerstore().AddAddrs(ha1.ID(), ha1.Addrs(), peerstore.PermanentAddrTTL)

			_ = NewCBORService(ha1, ch, nil)
			cli := NewCBORClient(ha2, ha1.ID(), nil)

			defer ha1.Close()
			defer ha2.Close()

			wg := &sync.WaitGroup{}
			// "creating stream to 12D3K***VtZcM6, stream-55822: transient: cannot reserve outbound stream: resource limit exceeded"
			total := 10
			// if pa[0] == "udp/23575/quic" {
			// 	total = 1000
			// }

			for i := 0; i < total; i++ {
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

					time.Sleep(time.Duration(rand.Int63n(int64(x%999) + 1)))
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
		})
	}
}

func makeBasicHost(paddr string) (host.Host, error) {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 0, nil)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/" + paddr),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
	}

	return libp2p.New(opts...)
}
