package main

import (
	"context"
	"fmt"
	"go/build"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/avalanche-network-runner/local"
	"github.com/ava-labs/avalanche-network-runner/network"
	"github.com/ava-labs/avalanchego/utils/logging"
)

const (
	healthyTimeout = 2 * time.Minute
)

// NewDefaultNetwork returns a new network using a pre-defined network configuration. The following addresses are pre-funded:
// X-Chain Address 1: X-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p
// X-Chain Address 1 Key: PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN
// X-Chain Address 2: X-custom16045mxr3s2cjycqe2xfluk304xv3ezhkhsvkpr
// X-Chain Address 2 Key: PrivateKey-2fzYBh3bbWemKxQmMfX6DSuL2BFmDSLQWTvma57xwjQjtf8gFq
// P-Chain Address 1: P-custom18jma8ppw3nhx5r4ap8clazz0dps7rv5u9xde7p
// P-Chain Address 1 Key: PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN
// P-Chain Address 2: P-custom16045mxr3s2cjycqe2xfluk304xv3ezhkhsvkpr
// P-Chain Address 2 Key: PrivateKey-2fzYBh3bbWemKxQmMfX6DSuL2BFmDSLQWTvma57xwjQjtf8gFq
// C-Chain Address 1: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC
// C-Chain Address 1 Key: 56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027
// C-Chain Address 2: 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641
// C-Chain Address 2 Key: dc3b75ce8741f4ae37b21c8659c28d0842cbd453b00d6b69adc8c34dae3a7644
//
// The following nodes are validators:
// * NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg
// * NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ
// * NodeID-NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN
// * NodeID-GWPcbFJZFfZreETSoWjPimr846mXEKCtu
// * NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5

var goPath = os.ExpandEnv("$GOPATH")

// Blocks until a signal is received on [signalChan], upon which
// [n.Stop()] is called. If [signalChan] is closed, does nothing.
// Closes [closedOnShutdownChan] amd [signalChan] when done shutting down network.
// This function should only be called once.
func shutdownOnSignal(
	log logging.Logger,
	n network.Network,
	signalChan chan os.Signal,
	closedOnShutdownChan chan struct{},
) {
	sig := <-signalChan
	log.Info("got OS signal %s", sig)
	if err := n.Stop(context.Background()); err != nil {
		log.Debug("error while stopping network: %s", err)
	}
	signal.Reset()
	close(signalChan)
	close(closedOnShutdownChan)
}

// Shows example usage of the Avalanche Network Runner.
// Creates a local five node Avalanche network
// and waits for all nodes to become healthy.
// The network runs until the user provides a SIGINT or SIGTERM.
func main() {
	// Create the logger
	cfg := logging.DefaultConfig
	cfg.DisplayLevel = logging.Debug
	logFactory := logging.NewFactory(cfg)
	log, err := logFactory.Make("main")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if goPath == "" {
		goPath = build.Default.GOPATH
	}
	binaryPath := fmt.Sprintf("%s%s", goPath, "/github.com/ava-labs/avalanchego/build/avalanchego")
	homeDir := os.ExpandEnv("$HOME")
	dbPath := fmt.Sprintf("%s/.%s/db", homeDir, "avalanchego")
	if err := run(log, binaryPath, dbPath); err != nil {
		log.Fatal("%s", err)
		os.Exit(1)
	}
}

func run(log logging.Logger, binaryPath, dbPath string) error {
	// Create the network
	cfg := local.NewDefaultConfig(binaryPath)
	if cfg.Flags == nil {
		cfg.Flags = make(map[string]interface{})
	}
	cfg.Flags["whitelisted-subnets"] = "k2N2VSSafGrN39VtDgMBdeson8mixzKk2k1ECBdfA3QezRe7G"

	nw, err := local.NewNetwork(log, cfg, dbPath)
	if err != nil {
		return err
	}
	defer func() { // Stop the network when this function returns
		if err := nw.Stop(context.Background()); err != nil {
			log.Debug("error stopping network: %w", err)
		}
	}()

	// When we get a SIGINT or SIGTERM, stop the network and close [closedOnShutdownCh]
	signalsChan := make(chan os.Signal, 1)
	signal.Notify(signalsChan, syscall.SIGINT)
	signal.Notify(signalsChan, syscall.SIGTERM)
	closedOnShutdownCh := make(chan struct{})
	go func() {
		shutdownOnSignal(log, nw, signalsChan, closedOnShutdownCh)
	}()

	// Wait until the nodes in the network are ready
	ctx, cancel := context.WithTimeout(context.Background(), healthyTimeout)
	defer cancel()
	healthyChan := nw.Healthy(ctx)
	log.Info("waiting for all nodes to report healthy...")
	if err := <-healthyChan; err != nil {
		return err
	}

	log.Info("All nodes healthy. Network will run until you CTRL + C to exit...")
	// Wait until done shutting down network after SIGINT/SIGTERM
	<-closedOnShutdownCh
	return nil
}
