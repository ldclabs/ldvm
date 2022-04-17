// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"fmt"
	"net/http"

	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ldclabs/ldvm/ld"
)

// {tx} or {unsignedTx, signatures[, exSignatures]}
type IssueTxArgs struct {
	Tx           string   `json:"tx"`
	UnsignedTx   string   `json:"unsignedTx"`
	Signatures   []string `json:"signatures"`
	ExSignatures []string `json:"exSignatures"`
}

// IssueTx
func (api *BlockChainAPI) IssueTx(_ *http.Request, args *IssueTxArgs, reply *GetReply) error {
	tx := &ld.Transaction{}
	if len(args.Tx) > 0 {
		data, err := formatting.Decode(formatting.CB58, args.Tx)
		if err != nil {
			return fmt.Errorf("invalid tx, CB58 decode failed: %v", err)
		}
		if err = tx.Unmarshal(data); err != nil {
			return fmt.Errorf("invalid tx, unmarshal failed: %v", err)
		}
	} else {
		data, err := formatting.Decode(formatting.CB58, args.UnsignedTx)
		if err != nil {
			return fmt.Errorf("invalid tx, CB58 decode failed: %v", err)
		}
		if err = tx.Unmarshal(data); err != nil {
			return fmt.Errorf("invalid tx, unmarshal failed: %v", err)
		}
		tx.Signatures, err = ld.SignaturesFromStrings(args.Signatures)
		if err != nil {
			return fmt.Errorf("invalid tx signatures: %v", err)
		}
		tx.ExSignatures, err = ld.SignaturesFromStrings(args.ExSignatures)
		if err != nil {
			return fmt.Errorf("invalid tx signatures: %v", err)
		}
	}

	if err := api.state.SubmitTx(tx); err != nil {
		return err
	}
	reply.ID = tx.ID().String()
	reply.Status = tx.Status.String()
	return nil
}
