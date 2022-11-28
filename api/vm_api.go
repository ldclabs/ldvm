// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"net/http"

	"github.com/ldclabs/ldvm/genesis"
)

type VMAPI struct {
	name, version string
}

func NewVMAPI(name, version string) *VMAPI {
	return &VMAPI{name, version}
}

type NoArgs struct{}

// Version returns VM's name and version
func (api *VMAPI) Version(_ *http.Request, args *NoArgs, reply *map[string]string) error {
	*reply = map[string]string{
		"name":    api.name,
		"version": api.version,
	}
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
