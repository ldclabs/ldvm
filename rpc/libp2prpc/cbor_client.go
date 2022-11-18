// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package libp2prpc

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/rs/xid"

	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/util/encoding"
)

const (
	ProtocolID  protocol.ID = "/cborrpc/v1"
	ServiceName string      = "ldclabs:cborrpc"
)

type CBORClient struct {
	host     host.Host
	endpoint peer.ID
}

func NewCBORClient(host host.Host, endpoint peer.ID) *CBORClient {
	return &CBORClient{
		host:     host,
		endpoint: endpoint,
	}
}

func (c *CBORClient) Request(ctx context.Context, method string, params, result interface{}) *cborrpc.Response {
	var err error

	req := &cborrpc.Request{Method: method}
	req.Params, err = encoding.MarshalCBOR(params)
	if err != nil {
		return req.Error(&cborrpc.Error{
			Code:    cborrpc.CodeInvalidParams,
			Message: err.Error(),
		})
	}

	res := c.Do(ctx, req)
	if result != nil {
		res.DecodeResult(result)
	}

	return res
}

func (c *CBORClient) Do(ctx context.Context, req *cborrpc.Request) *cborrpc.Response {
	err := ctx.Err()
	if err != nil {
		return req.Error(err)
	}

	if req.Method == "" {
		return req.InvalidMethod()
	}

	if req.ID == "" {
		req.ID = xid.New().String()
	}

	data, err := encoding.MarshalCBOR(req)
	if err != nil {
		return req.Error(err)
	}

	s, err := c.host.NewStream(ctx, c.endpoint, ProtocolID)
	if err != nil {
		return req.Error(fmt.Errorf("creating stream to %s, %v", c.endpoint, err))
	}
	defer s.Close()

	res := &cborrpc.Response{ID: req.ID}
	_, err = s.Write(data)
	if err != nil {
		return req.Error(fmt.Errorf("write data failed, %v", err))
	}
	s.CloseWrite()

	_, err = res.ReadFrom(s)
	if err != nil {
		return req.Error(err)
	}

	return res
}
