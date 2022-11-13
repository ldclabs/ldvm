// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ids"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/encoding"
)

type VMAPI struct{}

func NewVMAPI() *VMAPI {
	return &VMAPI{}
}

type NoArgs struct{}

type TxData interface {
	SyntacticVerify() error
	Bytes() []byte
}

type EncodeReply struct {
	Bytes    string              `json:"bytes"`
	Encoding formatting.Encoding `json:"encoding"` // hex
	Length   int                 `json:"length"`
}

func (r *EncodeReply) SetTxData(tx TxData) error {
	if err := tx.SyntacticVerify(); err != nil {
		return fmt.Errorf("invalid args: %v", err)
	}

	data := tx.Bytes()
	s, err := formatting.Encode(formatting.HexC, data)
	if err != nil {
		return fmt.Errorf("invalid args: %v", err)
	}
	r.Encoding = formatting.Hex
	r.Length = len(data)
	r.Bytes = s
	return nil
}

// Genesis returns the genesis data
func (api *VMAPI) Genesis(_ *http.Request, args *NoArgs, reply *genesis.Genesis) error {
	gs, err := genesis.FromJSON([]byte(genesis.LocalGenesisConfigJSON))
	if err != nil {
		return err
	}
	*reply = *gs
	return nil
}

type TransactionArgs struct {
	Type      uint16           `json:"type"`
	ChainID   uint64           `json:"chainID"`
	Nonce     uint64           `json:"nonce"`
	GasTip    uint64           `json:"gasTip"`
	GasFeeCap uint64           `json:"gasFeeCap"`
	From      ids.Address      `json:"from"`
	To        *ids.Address     `json:"to"`
	Token     *ids.TokenSymbol `json:"token"`
	Amount    *big.Int         `json:"amount"`
	Data      json.RawMessage  `json:"data"`
}

func (api *VMAPI) EncodeTx(_ *http.Request, args *TransactionArgs, reply *EncodeReply) error {
	data := encoding.UnmarshalJSONData(args.Data)
	tx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TxType(args.Type),
		ChainID:   args.ChainID,
		Nonce:     args.Nonce,
		GasTip:    args.GasTip,
		GasFeeCap: args.GasFeeCap,
		From:      args.From,
		To:        args.To,
		Token:     args.Token,
		Amount:    args.Amount,
		Data:      data,
	}}
	return reply.SetTxData(tx)
}
