// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txpool

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/cose/cose"
	"github.com/ldclabs/cose/cwt"
	"github.com/ldclabs/cose/iana"
	"github.com/ldclabs/cose/key"
	"github.com/ldclabs/cose/key/ed25519"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/rpc/protocol/cborrpc"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
	"github.com/ldclabs/ldvm/util/value"
)

type RequestParams struct {
	Payload cbor.RawMessage                `cbor:"p,omitempty"`
	CWT     *cose.Sign1Message[cwt.Claims] `cbor:"t,omitempty"`
}

func (p *TxPool) ServeRPC(ctx context.Context, req *cborrpc.Request) *cborrpc.Response {
	params := &RequestParams{}
	if err := req.DecodeParams(params); err != nil {
		return req.Error(err)
	}

	value.DoIfCtxValueValid(ctx, func(log *value.Log) {
		log.Set("rpcId", value.String(req.ID))
		log.Set("rpcMethod", value.String(req.Method))
		if params.CWT != nil {
			log.Set("subject", value.String(params.CWT.Payload.Subject))
		}
	})

	switch req.Method {
	case "sizeToBuild":
		return req.Result(p.rpcSizeToBuild(ctx))

	case "submitTxs":
		txs := &ld.Txs{}
		if err := txs.Unmarshal(params.Payload); err != nil {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: err.Error(),
			})
		}

		if err := p.rpcSubmitTxs(ctx, *txs); err != nil {
			return req.Error(err)
		}

		value.DoIfCtxValueValid(ctx, func(log *value.Log) {
			log.Set("txIDs", txs.IDs().ToValue())
		})
		return req.Result(true)

	case "loadByIDs":
		txIDs := ids.IDList[ids.ID32]{}
		if err := encoding.UnmarshalCBOR(params.Payload, &txIDs); err != nil {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: err.Error(),
			})
		}
		if err := txIDs.Valid(); err != nil {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: err.Error(),
			})
		}

		objs, err := p.rpcLoadByIDs(ctx, txIDs)
		if err != nil {
			return req.Error(err)
		}
		return req.Result(objs.ToRawList())

	case "fetchToBuild":
		if err := p.auth(params); err != nil {
			return req.Error(err)
		}

		amount := 0
		if err := encoding.UnmarshalCBOR(params.Payload, &amount); err != nil || amount < 1 {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: err.Error(),
			})
		}
		txs, err := p.rpcFetchToBuild(ctx, amount)
		if err != nil {
			return req.Error(err)
		}

		return req.Result(txs)

	case "updateBuildStatus":
		if err := p.auth(params); err != nil {
			return req.Error(err)
		}
		ts := &TxsBuildStatus{}
		if err := encoding.UnmarshalCBOR(params.Payload, ts); err != nil {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: err.Error(),
			})
		}
		err := p.rpcUpdateBuildStatus(ctx, ts)
		if err != nil {
			return req.Error(err)
		}
		return req.Result(true)

	case "acceptByBlock":
		if err := p.auth(params); err != nil {
			return req.Error(err)
		}
		blk := &ld.Block{}
		if err := encoding.UnmarshalCBOR(params.Payload, blk); err != nil {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: err.Error(),
			})
		}
		if err := blk.SyntacticVerify(); err != nil {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: err.Error(),
			})
		}

		if blk.Builder.String() != p.builder {
			return req.Error(&cborrpc.Error{
				Code:    cborrpc.CodeInvalidParams,
				Message: fmt.Sprintf("block builder %s does not match, expected %s", blk.Builder.String(), p.builder),
			})
		}

		err := p.rpcAcceptByBlock(ctx, blk)
		if err != nil {
			return req.Error(err)
		}
		return req.Result(true)

	case "updateBuilderKeepers":
		if err := p.auth(params); err != nil {
			return req.Error(err)
		}

		err := p.rpcUpdateBuilderKeepers(ctx)
		if err != nil {
			return req.Error(err)
		}

		return req.Result(true)

	default:
		return req.InvalidMethod()
	}
}

func (p *TxPool) auth(reqParams *RequestParams) error {
	if reqParams.CWT == nil {
		return &erring.Error{
			Code:    cborrpc.CodeInvalidRequest,
			Message: "txpool: no CWT",
		}
	}

	if reqParams.CWT.Payload.Audience != p.cwtAudience {
		return &erring.Error{
			Code:    cborrpc.CodeInvalidRequest,
			Message: "txpool: invalid CWT audience",
		}
	}

	if reqParams.CWT.Payload.Subject != p.builder {
		return &erring.Error{
			Code:    cborrpc.CodeInvalidRequest,
			Message: "txpool: invalid CWT subject",
		}
	}

	if !bytes.Equal(reqParams.CWT.Payload.CWTID, ids.ID32FromData(reqParams.Payload).Bytes()) {
		return &erring.Error{
			Code:    cborrpc.CodeInvalidRequest,
			Message: "txpool: invalid CWT ID",
		}
	}

	kid, err := reqParams.CWT.Unprotected.GetBytes(iana.HeaderParameterKid)
	if err != nil {
		return &erring.Error{
			Code:    cborrpc.CodeInvalidRequest,
			Message: "txpool: invalid CWT key ID",
		}
	}
	verifier := FindCOSEVerifier(p.builderKeepers, kid)
	if verifier == nil {
		return &erring.Error{
			Code:    cborrpc.CodeServerError - http.StatusUnauthorized,
			Message: "txpool: signatures verify failed",
		}
	}
	if err := reqParams.CWT.Verify(verifier, p.cwtExData); err != nil {
		return &erring.Error{
			Code:    cborrpc.CodeServerError - http.StatusUnauthorized,
			Message: "txpool: " + err.Error(),
		}
	}

	return nil
}

func FindCOSEVerifier(ks signer.Keys, kid []byte) key.Verifier {
	for _, k := range ks {
		if bytes.Equal(k.Address().Bytes(), kid) && k.Kind() == signer.Ed25519 {
			pubKey, err := ed25519.KeyFromPublic(k.Bytes())
			if err != nil {
				return nil
			}
			verifier, _ := pubKey.Verifier()
			return verifier
		}
	}
	return nil
}
