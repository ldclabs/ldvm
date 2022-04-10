// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

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
	"github.com/ldclabs/ldvm/ld"
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
func (v *VM) Initialize(
	ctx *snow.Context,
	dbManager manager.Manager,
	genesisData []byte,
	upgradeData []byte,
	configData []byte,
	toEngine chan<- common.Message,
	_ []*common.Fx,
	appSender common.AppSender,
) (err error) {
	v.ctx = ctx
	v.dbManager = dbManager
	v.toEngine = toEngine
	v.appSender = appSender

	var cfg *config.Config
	cfg, err = config.New(configData)
	if err != nil {
		return fmt.Errorf("LDVM Initialize failed to get config: %s", err)
	}

	cfg.Logger.MsgPrefix = fmt.Sprintf("%s@%s", Name, Version)
	logFactory := logging.NewFactory(cfg.Logger)
	v.Log, err = logFactory.Make("ldvm-" + ctx.NodeID.String())
	if err != nil {
		return fmt.Errorf("LDVM Initialize failed to create logger: %s", err)
	}

	v.Log.Info("LDVM Initialize with genesisData: <%s>", string(genesisData))
	v.Log.Info("LDVM Initialize with upgradeData: <%s>", string(upgradeData))
	v.Log.Info("LDVM Initialize with configData: <%s>", string(configData))

	err = ld.Recover("LDVM Initialize", func() error {
		return v.initialize(cfg, genesisData)
	})
	if err != nil {
		v.Log.Error(err.Error())
	}
	return err
}

func (v *VM) initialize(cfg *config.Config, genesisData []byte) (err error) {
	var gs *genesis.Genesis
	if len(genesisData) == 0 {
		genesisData = []byte(genesis.LocalGenesisConfigJSON)
	}

	gs, err = genesis.FromJSON(genesisData)
	if err != nil {
		return fmt.Errorf("parse genesis data error: %v", err)
	}
	v.state = chain.NewState(v.ctx, v.dbManager.Current().Database, gs, cfg, v.Log)
	if err = v.state.Bootstrap(); err != nil {
		return fmt.Errorf("bootstrap error: %v", err)
	}
	return nil
}

// SetState implements the common.VM SetState interface
// SetState communicates to VM its next state it starts
func (v *VM) SetState(state snow.State) error {
	v.Log.Info("SetState %v", state)
	return v.state.SetState(state)
}

// Shutdown implements the common.VM Shutdown interface
// Shutdown is called when the node is shutting down.
func (v *VM) Shutdown() error {
	v.Log.Info("Shutdown")
	v.dbManager.Close()
	return nil
}

// Version implements the common.VM Version interface
// Version returns the version of the VM this node is running.
func (v *VM) Version() (string, error) {
	return Version.String(), nil
}

// CreateStaticHandlers implements the common.VM CreateStaticHandlers interface
// Creates the HTTP handlers for custom VM network calls.
//
// This exposes handlers that the outside world can use to communicate with
// a static reference to the VM. Each handler has the path:
// [Address of node]/ext/vm/[VM ID]/[extension]
func (v *VM) CreateStaticHandlers() (map[string]*common.HTTPHandler, error) {
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
func (v *VM) CreateHandlers() (map[string]*common.HTTPHandler, error) {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	if err := server.RegisterService(api.NewBlockChainAPI(v.state), Name); err != nil {
		v.Log.Error("CreateHandlers error: %v", err)
		return nil, err
	}

	v.Log.Error("CreateHandlers ok")
	return map[string]*common.HTTPHandler{
		"/rpc": {
			LockOptions: common.WriteLock,
			Handler:     server,
		},
	}, nil
}

// HealthCheck implements the common.VM health.Checker HealthCheck interface
// Returns nil if the VM is healthy.
// Periodically called and reported via the node's Health API.
func (v *VM) HealthCheck() (interface{}, error) {
	return v.state.HealthCheck()
}

// Connected implements the common.VM validators.Connector Connected interface
// Connector represents a handler that is called when a connection is marked as connected
func (v *VM) Connected(id ids.ShortID, nodeVersion version.Application) error {
	v.Log.Info("Connected %s, %v", id, nodeVersion)
	return nil // noop
}

// Disconnected implements the common.VM Disconnected interface
// Connector represents a handler that is called when a connection is marked as disconnected
func (v *VM) Disconnected(id ids.ShortID) error {
	v.Log.Info("Disconnected %s", id)
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
func (v *VM) AppGossip(nodeID ids.ShortID, msg []byte) error {
	v.Log.Info("AppGossip %s, %d bytes", nodeID, len(msg))
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
func (v *VM) AppRequest(nodeID ids.ShortID, requestID uint32, time time.Time, request []byte) error {
	v.Log.Info("AppRequest %s, %d, %d bytes", nodeID, requestID, len(request))
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
func (v *VM) AppResponse(nodeID ids.ShortID, requestID uint32, response []byte) error {
	v.Log.Info("AppResponse %s, %d, %d bytes", nodeID, requestID, len(response))
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
func (v *VM) AppRequestFailed(nodeID ids.ShortID, requestID uint32) error {
	v.Log.Info("AppRequestFailed %s, %d", nodeID, requestID)
	return nil
}

// GetBlock implements the block.ChainVM GetBlock interface
//
// GetBlock attempt to fetch a block by it's ID
// If the block does not exist, an error should be returned.
func (v *VM) GetBlock(id ids.ID) (blk snowman.Block, err error) {
	blk, err = v.state.GetBlock(id)
	if err != nil {
		v.Log.Error("GetBlock %s error: %v", id, err)
	} else {
		v.Log.Info("GetBlock %s", blk.ID())
	}
	return
}

// ParseBlock implements the block.ChainVM ParseBlock interface
//
// ParseBlock attempt to fetch a block by its bytes.
func (v *VM) ParseBlock(data []byte) (blk snowman.Block, err error) {
	err = ld.Recover("", func() error {
		blk, err = v.state.ParseBlock(data)
		return err
	})

	if err != nil {
		v.Log.Error("ParseBlock %v", err)
	} else {
		v.Log.Info("ParseBlock %s", blk.ID())
	}
	return
}

// BuildBlock implements the block.ChainVM BuildBlock interface
//
// BuildBlock attempt to create a new block from data contained in the VM.
// If the VM doesn't want to issue a new block, an error should be returned.
func (v *VM) BuildBlock() (blk snowman.Block, err error) {
	blk, err = v.state.BuildBlock()
	if err != nil {
		v.Log.Error("BuildBlock %v", err)
	} else {
		v.Log.Info("BuildBlock %s", blk.ID())
	}
	return
}

// SetPreference implements the block.ChainVM SetPreference interface
//
// SetPreference notify the VM of the currently preferred block.
// This should always be a block that has no children known to consensus.
func (v *VM) SetPreference(id ids.ID) error {
	err := v.state.SetPreference(id)
	if err != nil {
		v.Log.Error("SetPreference %s error %v", id, err)
	} else {
		v.Log.Info("SetPreference %s", id)
	}
	return err
}

// LastAccepted implements the block.ChainVM LastAccepted interface
//
// LastAccepted returns the ID of the last accepted block.
// If no blocks have been accepted by consensus yet, it is assumed there is
// a definitionally accepted block, the Genesis block, that will be
// returned.
func (v *VM) LastAccepted() (ids.ID, error) {
	block, err := v.state.LastAcceptedBlock()

	if err != nil {
		v.Log.Error("LastAccepted %v", err)
		return ids.Empty, err
	}

	v.Log.Info("LastAccepted %s", block.ID())
	return block.ID(), nil
}

// VerifyHeightIndex implements the block.HeightIndexedChainVM VerifyHeightIndex interface
// VerifyHeightIndex should return:
// - nil if the height index is available.
// - ErrHeightIndexedVMNotImplemented if the height index is not supported.
// - ErrIndexIncomplete if the height index is not currently available.
// - Any other non-standard error that may have occurred when verifying the index.
func (v *VM) VerifyHeightIndex() error {
	return nil
}

// GetBlockIDAtHeight implements the block.HeightIndexedChainVM GetBlockIDAtHeight interface
// GetBlockIDAtHeight returns the ID of the block that was accepted with [height].
func (v *VM) GetBlockIDAtHeight(height uint64) (ids.ID, error) {
	id, err := v.state.GetBlockIDAtHeight(height)
	if err != nil {
		v.Log.Error("GetBlockIDAtHeight %d error %v", height, err)
	} else {
		v.Log.Info("GetBlockIDAtHeight %d: %s", height, id)
	}
	return id, err
}
