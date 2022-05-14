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
	"github.com/ldclabs/ldvm/util"

	"github.com/ldclabs/ldvm/ld"
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
	s, err := formatting.EncodeWithChecksum(formatting.Hex, data)
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
	Type      uint8           `json:"type"`
	ChainID   uint64          `json:"chainID"`
	Nonce     uint64          `json:"Nonce"`
	GasTip    uint64          `json:"gasTip"`
	GasFeeCap uint64          `json:"gasFeeCap"`
	From      util.EthID      `json:"from"`
	To        *util.EthID     `json:"to"`
	Amount    *big.Int        `json:"amount"`
	Data      json.RawMessage `json:"data"`
}

func (api *VMAPI) EncodeTx(_ *http.Request, args *TransactionArgs, reply *EncodeReply) error {
	data := util.JSONUnmarshalData(args.Data)
	tx := &ld.Transaction{
		Type:      ld.TxType(args.Type),
		ChainID:   args.ChainID,
		Nonce:     args.Nonce,
		GasTip:    args.GasTip,
		GasFeeCap: args.GasFeeCap,
		From:      args.From,
		To:        args.To,
		Amount:    args.Amount,
		Data:      data,
	}
	return reply.SetTxData(tx)
}

type DataMetaArgs struct {
	ModelID   util.ModelID    `json:"mID"`
	Version   uint64          `json:"version"`
	Threshold uint8           `json:"threshold"`
	Keepers   []util.EthID    `json:"keepers"`
	Data      json.RawMessage `json:"data"`
}

func (api *VMAPI) EncodeCreateData(_ *http.Request, args *DataMetaArgs, reply *EncodeReply) error {
	data := util.JSONUnmarshalData(args.Data)
	tx := &ld.DataMeta{
		ModelID:   args.ModelID,
		Version:   args.Version,
		Threshold: args.Threshold,
		Keepers:   args.Keepers,
		Data:      data,
	}
	return reply.SetTxData(tx)
}

type ModelMetaArgs struct {
	Name      string          `json:"name"`
	Threshold uint8           `json:"threshold"`
	Keepers   []util.EthID    `json:"keepers"`
	Data      json.RawMessage `json:"data"`
}

func (api *VMAPI) EncodeCreateModel(_ *http.Request, args *ModelMetaArgs, reply *EncodeReply) error {
	data := util.JSONUnmarshalData(args.Data)
	tx := &ld.ModelMeta{
		Name:      args.Name,
		Threshold: args.Threshold,
		Keepers:   args.Keepers,
		Data:      data,
	}
	return reply.SetTxData(tx)
}

// func (api *VMAPI) EncodeName(_ *http.Request, args *service.Name, reply *EncodeReply) error {
// 	return reply.SetTxData(args)
// }

// func (api *VMAPI) EncodeInfo(_ *http.Request, args *service.Profile, reply *EncodeReply) error {
// 	return reply.SetTxData(args)
// }
