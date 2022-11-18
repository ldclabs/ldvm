// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"fmt"
	"math/rand"
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

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
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

func TestCBORRPC(t *testing.T) {
	assert := assert.New(t)

	ch := &cborhandler{snap: true}
	ha1, err := makeBasicHost(13131, false)
	require.NoError(t, err)

	ha2, err := makeBasicHost(13132, false)
	require.NoError(t, err)
	ha2.Peerstore().AddAddrs(ha1.ID(), ha1.Addrs(), peerstore.PermanentAddrTTL)

	_ = NewCBORService(ha1, ch)
	cli := NewCBORClient(ha2, ha1.ID())

	defer ha1.Close()
	defer ha2.Close()

	t.Run("Request should work", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		re := &result{}
		res := cli.Request(ctx, "TestMethod", 1234, re)
		require.Nil(t, res.Error)

		assert.Nil(ch.err)
		assert.Equal("TestMethod", re.Method)
		assert.Equal(encoding.MustMarshalCBOR(1234), []byte(re.Params))
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
	ha1, err := makeBasicHost(13131, false)
	require.NoError(t, err)
	ha1.Network().ResourceManager().Close()

	ha2, err := makeBasicHost(13132, false)
	require.NoError(t, err)
	ha2.Peerstore().AddAddrs(ha1.ID(), ha1.Addrs(), peerstore.PermanentAddrTTL)

	_ = NewCBORService(ha1, ch)
	cli := NewCBORClient(ha2, ha1.ID())

	defer ha1.Close()
	defer ha2.Close()

	wg := &sync.WaitGroup{}
	// "creating stream to 12D3K***VtZcM6, stream-55822: transient: cannot reserve outbound stream: resource limit exceeded"
	for i := 0; i < 100; i++ {
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

func makeBasicHost(listenPort int, insecure bool) (host.Host, error) {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 0, nil)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
	}

	if insecure {
		opts = append(opts, libp2p.NoSecurity)
	}

	return libp2p.New(opts...)
}
