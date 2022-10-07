// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"fmt"
	"net/http"

	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/signer"
)

type IssueTxArgs struct {
	Tx           string       `json:"tx"`
	Signatures   []signer.Sig `json:"sigs"`
	ExSignatures []signer.Sig `json:"exSigs"`
}

// IssueTx
func (api *BlockChainAPI) IssueTx(_ *http.Request, args *IssueTxArgs, reply *GetReply) error {
	tx := &ld.Transaction{}

	data, err := decodeBytes(args.Tx)
	if err != nil {
		return fmt.Errorf("invalid tx, Hex decode failed: %v", err)
	}
	if err = tx.Unmarshal(data); err != nil {
		return fmt.Errorf("invalid tx, unmarshal failed: %v", err)
	}

	if len(args.Signatures) > 0 {
		tx.Signatures = args.Signatures
	}

	if len(args.ExSignatures) > 0 {
		tx.ExSignatures = args.ExSignatures
	}

	if err := api.bc.SubmitTx(tx); err != nil {
		return err
	}

	reply.ID = tx.ID.String()
	reply.Status = choices.Unknown.String()
	return nil
}
