// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"net/http"

	"github.com/ipld/go-ipld-prime/datamodel"
)

type VMAPI struct{}

func NewVMAPI() *VMAPI {
	return &VMAPI{}
}

// EncodeArgs are arguments for Encode
type EncodeArgs struct {
	Data datamodel.Node `json:"data"`
}

// EncodeReply is the reply from Encode
type EncodeReply struct {
	Bytes string `json:"bytes"`
}

// Encode returns the encoded data
func (api *VMAPI) Encode(_ *http.Request, args *EncodeArgs, reply *EncodeReply) error {
	return nil
}

// DecodeArgs are arguments for Decode
type DecodeArgs struct {
	Bytes string `json:"bytes"`
}

// DecodeReply is the reply from Decode
type DecodeReply struct {
	Data datamodel.Node `json:"data"`
}

// Decode returns the Decoded data
func (api *VMAPI) Decode(_ *http.Request, args *DecodeArgs, reply *DecodeReply) error {
	return nil
}
