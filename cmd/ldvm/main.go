// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/ulimit"
	"github.com/ava-labs/avalanchego/vms/rpcchainvm"
	"github.com/hashicorp/go-plugin"

	"github.com/ldclabs/ldvm/ldvm"
)

var version = flag.Bool("version", false, "show LDVM version")

func main() {
	// Print VM ID and exit
	if *version {
		fmt.Printf("%s@%s\n", ldvm.Name, ldvm.Version)
		os.Exit(0)
	}

	if err := ulimit.Set(ulimit.DefaultFDLimit, (*logging.Log)(nil)); err != nil {
		fmt.Printf("failed to set fd limit correctly due to: %s", err)
		os.Exit(1)
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: rpcchainvm.Handshake,
		Plugins: map[string]plugin.Plugin{
			"vm": rpcchainvm.New(&ldvm.VM{}),
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
