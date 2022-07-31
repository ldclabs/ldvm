// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"fmt"
	"net/http"

	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
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
		data, err := decodeBytes(args.Tx)
		if err != nil {
			return fmt.Errorf("invalid tx, Hex decode failed: %v", err)
		}
		if err = tx.Unmarshal(data); err != nil {
			return fmt.Errorf("invalid tx, unmarshal failed: %v", err)
		}
	} else {
		data, err := decodeBytes(args.UnsignedTx)
		if err != nil {
			return fmt.Errorf("invalid tx, Hex decode failed: %v", err)
		}
		if err = tx.Unmarshal(data); err != nil {
			return fmt.Errorf("invalid tx, unmarshal failed: %v", err)
		}
		tx.Signatures, err = util.SignaturesFromStrings(args.Signatures)
		if err != nil {
			return fmt.Errorf("invalid tx signatures: %v", err)
		}
		tx.ExSignatures, err = util.SignaturesFromStrings(args.ExSignatures)
		if err != nil {
			return fmt.Errorf("invalid tx signatures: %v", err)
		}
	}

	if err := api.bc.SubmitTx(tx); err != nil {
		return err
	}
	reply.ID = tx.ID.String()
	reply.Status = choices.Unknown.String()
	return nil
}
