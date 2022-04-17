// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ldclabs/ldvm/genesis"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/app"
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
	Encoding formatting.Encoding `json:"encoding"` // cb58
	Length   int                 `json:"length"`
}

func (r *EncodeReply) SetTxData(tx TxData) error {
	if err := tx.SyntacticVerify(); err != nil {
		return fmt.Errorf("invalid args: %v", err)
	}
	data := tx.Bytes()
	s, err := formatting.EncodeWithChecksum(formatting.CB58, data)
	if err != nil {
		return fmt.Errorf("invalid args: %v", err)
	}
	r.Encoding = formatting.CB58
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
	From      ld.EthID        `json:"from"`
	To        ld.EthID        `json:"to"`
	Amount    *big.Int        `json:"amount"`
	Data      json.RawMessage `json:"data"`
}

func (api *VMAPI) EncodeTx(_ *http.Request, args *TransactionArgs, reply *EncodeReply) error {
	data := ld.JSONUnmarshalData(args.Data)
	tx := &ld.Transaction{
		Type:      ld.TxType(args.Type),
		ChainID:   args.ChainID,
		Nonce:     args.Nonce,
		GasTip:    args.GasTip,
		GasFeeCap: args.GasFeeCap,
		From:      ids.ShortID(args.From),
		To:        ids.ShortID(args.To),
		Amount:    args.Amount,
		Data:      data,
	}
	return reply.SetTxData(tx)
}

type DataMetaArgs struct {
	ModelID   ld.ModelID      `json:"modelID"`
	Version   uint64          `json:"version"`
	Threshold uint8           `json:"threshold"`
	Keepers   []ld.EthID      `json:"keepers"`
	Data      json.RawMessage `json:"data"`
}

func (api *VMAPI) EncodeCreateData(_ *http.Request, args *DataMetaArgs, reply *EncodeReply) error {
	data := ld.JSONUnmarshalData(args.Data)
	tx := &ld.DataMeta{
		ModelID:   ids.ShortID(args.ModelID),
		Version:   args.Version,
		Threshold: args.Threshold,
		Keepers:   ld.EthIDsToShort(args.Keepers...),
		Data:      data,
	}
	return reply.SetTxData(tx)
}

type ModelMetaArgs struct {
	Name      string          `json:"name"`
	Threshold uint8           `json:"threshold"`
	Keepers   []ld.EthID      `json:"keepers"`
	Data      json.RawMessage `json:"data"`
}

func (api *VMAPI) EncodeCreateModel(_ *http.Request, args *ModelMetaArgs, reply *EncodeReply) error {
	data := ld.JSONUnmarshalData(args.Data)
	tx := &ld.ModelMeta{
		Name:      args.Name,
		Threshold: args.Threshold,
		Keepers:   ld.EthIDsToShort(args.Keepers...),
		Data:      data,
	}
	return reply.SetTxData(tx)
}

type TxTransferArgs struct {
	Nonce  uint64          `json:"nonce"`
	From   ld.EthID        `json:"from"`
	To     ld.EthID        `json:"to"`
	Amount *big.Int        `json:"amount"`
	Expire uint64          `json:"expire"`
	Data   json.RawMessage `json:"data"`
}

func (api *VMAPI) EncodeTxTransfer(_ *http.Request, args *TxTransferArgs, reply *EncodeReply) error {
	data := ld.JSONUnmarshalData(args.Data)
	tx := &ld.TxTransfer{
		Nonce:  args.Nonce,
		From:   ids.ShortID(args.From),
		To:     ids.ShortID(args.To),
		Amount: args.Amount,
		Expire: args.Expire,
		Data:   data,
	}
	return reply.SetTxData(tx)
}

type TxUpdaterArgs struct {
	Version   uint64          `json:"version"`
	Threshold uint8           `json:"threshold"`
	Keepers   []ld.EthID      `json:"keepers"`
	To        ld.EthID        `json:"to"`
	Amount    *big.Int        `json:"amount"`
	Expire    uint64          `json:"expire"`
	Data      json.RawMessage `json:"data"`
}

func (api *VMAPI) EncodeTxUpdater(_ *http.Request, args *TxUpdaterArgs, reply *EncodeReply) error {
	data := ld.JSONUnmarshalData(args.Data)
	tx := &ld.TxUpdater{
		Version:   args.Version,
		Threshold: args.Threshold,
		Keepers:   ld.EthIDsToShort(args.Keepers...),
		To:        ids.ShortID(args.To),
		Amount:    args.Amount,
		Expire:    args.Expire,
		Data:      data,
	}
	return reply.SetTxData(tx)
}

func (api *VMAPI) EncodeName(_ *http.Request, args *app.Name, reply *EncodeReply) error {
	return reply.SetTxData(app.NewName(args))
}

func (api *VMAPI) EncodeInfo(_ *http.Request, args *app.Profile, reply *EncodeReply) error {
	return reply.SetTxData(app.NewProfile(args))
}
