// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"net/http"

	"github.com/ldclabs/ldvm/genesis"
)

type VMAPI struct{}

func NewVMAPI() *VMAPI {
	return &VMAPI{}
}

type NoArgs struct{}

// Genesis returns the genesis data
func (api *VMAPI) Genesis(_ *http.Request, args *NoArgs, reply *genesis.Genesis) error {
	gs, err := genesis.FromJSON([]byte(genesis.LocalGenesisConfigJSON))
	if err != nil {
		return err
	}
	*reply = *gs
	return nil
}
