// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ldvm

import (
	"fmt"
	"time"

	"github.com/gorilla/rpc/v2"

	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/version"

	"github.com/ldclabs/ldvm/api"
	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/genesis"
)

const (
	Name = "ldvm"
)

var (
	Version = version.NewDefaultVersion(1, 0, 0)

	_ block.ChainVM              = &VM{}
	_ block.HeightIndexedChainVM = &VM{}
)

// VM implements the snowman.VM interface
type VM struct {
	ctx       *snow.Context
	dbManager manager.Manager
	Log       logging.Logger

	// State of this VM
	state chain.StateDB

	// channel to send messages to the consensus engine
	toEngine  chan<- common.Message
	appSender common.AppSender
}

// Initialize implements the common.VM Initialize interface
// Initialize this VM.
// [ctx]: Metadata about this VM.
//     [ctx.networkID]: The ID of the network this VM's chain is running on.
//     [ctx.chainID]: The unique ID of the chain this VM is running on.
//     [ctx.Log]: Used to log messages
//     [ctx.NodeID]: The unique staker ID of this node.
//     [ctx.Lock]: A Read/Write lock shared by this VM and the consensus
//                 engine that manages this VM. The write lock is held
//                 whenever code in the consensus engine calls the VM.
// [dbManager]: The manager of the database this VM will persist data to.
// [genesisBytes]: The byte-encoding of the genesis information of this
//                 VM. The VM uses it to initialize its state. For
//                 example, if this VM were an account-based payments
//                 system, `genesisBytes` would probably contain a genesis
//                 transaction that gives coins to some accounts, and this
//                 transaction would be in the genesis block.
// [toEngine]: The channel used to send messages to the consensus engine.
// [fxs]: Feature extensions that attach to this VM.
func (vm *VM) Initialize(
	ctx *snow.Context,
	dbManager manager.Manager,
	genesisData []byte,
	upgradeData []byte,
	configData []byte,
	toEngine chan<- common.Message,
	_ []*common.Fx,
	appSender common.AppSender,
) error {
	vm.ctx = ctx
	vm.dbManager = dbManager
	vm.toEngine = toEngine
	vm.appSender = appSender

	logCfg := logging.DefaultConfig
	logFactory := logging.NewFactory(logCfg)
	log, err := logFactory.MakeChain(fmt.Sprintf("LDVM-%d", ctx.NetworkID))
	if err != nil {
		return fmt.Errorf("failed to create logger factory: %s", err)
	}
	vm.Log = log
	gs, err := genesis.FromJSON(genesisData)
	if err != nil {
		return err
	}

	cfg, err := config.New(configData)
	if err != nil {
		return err
	}

	vm.state = chain.NewState(ctx, dbManager.Current().Database, gs, cfg)
	if err := vm.state.Bootstrap(); err != nil {
		return err
	}
	return nil
}

// SetState implements the common.VM SetState interface
// SetState communicates to VM its next state it starts
func (vm *VM) SetState(state snow.State) error {
	return vm.state.SetState(state)
}

// Shutdown implements the common.VM Shutdown interface
// Shutdown is called when the node is shutting down.
func (vm *VM) Shutdown() error {
	vm.dbManager.Close()
	return nil
}

// Version implements the common.VM Version interface
// Version returns the version of the VM this node is running.
func (vm *VM) Version() (string, error) {
	return Version.String(), nil
}

// CreateStaticHandlers implements the common.VM CreateStaticHandlers interface
// Creates the HTTP handlers for custom VM network calls.
//
// This exposes handlers that the outside world can use to communicate with
// a static reference to the VM. Each handler has the path:
// [Address of node]/ext/vm/[VM ID]/[extension]
func (vm *VM) CreateStaticHandlers() (map[string]*common.HTTPHandler, error) {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	if err := server.RegisterService(api.NewVMAPI(), Name); err != nil {
		return nil, err
	}

	return map[string]*common.HTTPHandler{
		"": {
			LockOptions: common.NoLock,
			Handler:     server,
		},
	}, nil
}

// CreateHandlers implements the common.VM CreateHandlers interface
// Creates the HTTP handlers for custom chain network calls.
//
// This exposes handlers that the outside world can use to communicate with
// the chain. Each handler has the path:
// [Address of node]/ext/bc/[chain ID]/[extension]
func (vm *VM) CreateHandlers() (map[string]*common.HTTPHandler, error) {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	if err := server.RegisterService(api.NewBlockChainAPI(nil), Name); err != nil {
		return nil, err
	}

	return map[string]*common.HTTPHandler{
		"": {
			Handler: server,
		},
	}, nil
}

// HealthCheck implements the common.VM health.Checker HealthCheck interface
// Returns nil if the VM is healthy.
// Periodically called and reported via the node's Health API.
func (vm *VM) HealthCheck() (interface{}, error) { return nil, nil }

// Connected implements the common.VM validators.Connector Connected interface
// Connector represents a handler that is called when a connection is marked as connected
func (vm *VM) Connected(id ids.ShortID, nodeVersion version.Application) error {
	return nil // noop
}

// Disconnected implements the common.VM Disconnected interface
// Connector represents a handler that is called when a connection is marked as disconnected
func (vm *VM) Disconnected(id ids.ShortID) error {
	return nil // noop
}

// AppGossip implements the common.VM AppHandler AppGossip interface
// This VM doesn't (currently) have any app-specific messages
//
// Notify this engine of a gossip message from [nodeID].
//
// The meaning of [msg] is application (VM) specific, and the VM defines how
// to react to this message.
//
// This message is not expected in response to any event, and it does not
// need to be responded to.
//
// A node may gossip the same message multiple times. That is,
// AppGossip([nodeID], [msg]) may be called multiple times.
func (vm *VM) AppGossip(nodeID ids.ShortID, msg []byte) error {
	return nil
}

// AppRequest implements the common.VM AppRequest interface
// This VM doesn't (currently) have any app-specific messages
//
// Notify this engine of a request for data from [nodeID].
//
// The meaning of [request], and what should be sent in response to it, is
// application (VM) specific.
//
// It is not guaranteed that:
// * [request] is well-formed/valid.
//
// This node should typically send an AppResponse to [nodeID] in response to
// a valid message using the same request ID before the deadline. However,
// the VM may arbitrarily choose to not send a response to this request.
func (vm *VM) AppRequest(nodeID ids.ShortID, requestID uint32, time time.Time, request []byte) error {
	return nil
}

// AppResponse implements the common.VM AppResponse interface
// This VM doesn't (currently) have any app-specific messages
//
// Notify this engine of a response to the AppRequest message it sent to
// [nodeID] with request ID [requestID].
//
// The meaning of [response] is application (VM) specifc.
//
// It is guaranteed that:
// * This engine sent a request to [nodeID] with ID [requestID].
// * AppRequestFailed([nodeID], [requestID]) has not already been called.
// * AppResponse([nodeID], [requestID]) has not already been called.
//
// It is not guaranteed that:
// * [response] contains the expected response
// * [response] is well-formed/valid.
//
// If [response] is invalid or not the expected response, the VM chooses how
// to react. For example, the VM may send another AppRequest, or it may give
// up trying to get the requested information.
func (vm *VM) AppResponse(nodeID ids.ShortID, requestID uint32, response []byte) error {
	return nil
}

// AppRequestFailed implements the common.VM AppRequestFailed interface
// This VM doesn't (currently) have any app-specific messages
//
// Notify this engine that an AppRequest message it sent to [nodeID] with
// request ID [requestID] failed.
//
// This may be because the request timed out or because the message couldn't
// be sent to [nodeID].
//
// It is guaranteed that:
// * This engine sent a request to [nodeID] with ID [requestID].
// * AppRequestFailed([nodeID], [requestID]) has not already been called.
// * AppResponse([nodeID], [requestID]) has not already been called.
func (vm *VM) AppRequestFailed(nodeID ids.ShortID, requestID uint32) error {
	return nil
}

// GetBlock implements the block.ChainVM GetBlock interface
//
// GetBlock attempt to fetch a block by it's ID
// If the block does not exist, an error should be returned.
func (vm *VM) GetBlock(id ids.ID) (snowman.Block, error) {
	return vm.state.GetBlock(id)
}

// ParseBlock implements the block.ChainVM ParseBlock interface
//
// ParseBlock attempt to fetch a block by its bytes.
func (vm *VM) ParseBlock(data []byte) (snowman.Block, error) {
	return vm.state.ParseBlock(data)
}

// BuildBlock implements the block.ChainVM BuildBlock interface
//
// BuildBlock attempt to create a new block from data contained in the VM.
// If the VM doesn't want to issue a new block, an error should be returned.
func (vm *VM) BuildBlock() (snowman.Block, error) {
	return nil, nil
}

// SetPreference implements the block.ChainVM SetPreference interface
//
// SetPreference notify the VM of the currently preferred block.
// This should always be a block that has no children known to consensus.
func (vm *VM) SetPreference(id ids.ID) error {
	return vm.state.SetPreference(id)
}

// LastAccepted implements the block.ChainVM LastAccepted interface
//
// LastAccepted returns the ID of the last accepted block.
// If no blocks have been accepted by consensus yet, it is assumed there is
// a definitionally accepted block, the Genesis block, that will be
// returned.
func (vm *VM) LastAccepted() (ids.ID, error) {
	block, err := vm.state.LastAcceptedBlock()
	if err != nil {
		return ids.Empty, err
	}
	return block.ID(), nil
}

// VerifyHeightIndex implements the block.HeightIndexedChainVM VerifyHeightIndex interface
// VerifyHeightIndex should return:
// - nil if the height index is available.
// - ErrHeightIndexedVMNotImplemented if the height index is not supported.
// - ErrIndexIncomplete if the height index is not currently available.
// - Any other non-standard error that may have occurred when verifying the index.
func (vm *VM) VerifyHeightIndex() error {
	return nil
}

// GetBlockIDAtHeight implements the block.HeightIndexedChainVM GetBlockIDAtHeight interface
// GetBlockIDAtHeight returns the ID of the block that was accepted with [height].
func (vm *VM) GetBlockIDAtHeight(height uint64) (ids.ID, error) {
	return vm.state.GetBlockIDAtHeight(height)
}
