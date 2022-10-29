// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/rpc/v2"
	"go.uber.org/zap"

	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/utils/json"
	avalogging "github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/version"

	"github.com/ldclabs/ldvm/api"
	"github.com/ldclabs/ldvm/chain"
	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
)

const (
	Name = "ldvm"
)

var (
	Version = &version.Semantic{
		Major: 1,
		Minor: 0,
		Patch: 0,
	}

	_ block.ChainVM              = &VM{}
	_ block.HeightIndexedChainVM = &VM{}
)

// VM implements the snowman.VM interface
type VM struct {
	mu        sync.RWMutex
	ctx       *snow.Context
	dbManager manager.Manager
	Log       avalogging.Logger
	toEngine  chan<- common.Message
	appSender common.AppSender

	// State of this VM
	bc      chain.BlockChain
	network *PushNetwork
}

// Initialize implements the common.VM Initialize interface
// Initialize this VM.
// [ctx]: Metadata about this VM.
//
//	[ctx.networkID]: The ID of the network this VM's chain is running on.
//	[ctx.chainID]: The unique ID of the chain this VM is running on.
//	[ctx.Log]: Used to log messages
//	[ctx.NodeID]: The unique staker ID of this node.
//	[ctx.Lock]: A Read/Write lock shared by this VM and the consensus
//	            engine that manages this VM. The write lock is held
//	            whenever code in the consensus engine calls the VM.
//
// [dbManager]: The manager of the database this VM will persist data to.
// [genesisBytes]: The byte-encoding of the genesis information of this
//
//	VM. The VM uses it to initialize its state. For
//	example, if this VM were an account-based payments
//	system, `genesisBytes` would probably contain a genesis
//	transaction that gives coins to some accounts, and this
//	transaction would be in the genesis block.
//
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
	v.appSender = appSender
	v.toEngine = toEngine
	v.NewPushNetwork()

	errp := util.ErrPrefix("LDVM.Initialize error: ")
	var cfg *config.Config
	cfg, err = config.New(configData)
	if err != nil {
		return errp.Errorf("failed to get config, %v", err)
	}

	cfg.Logger.MsgPrefix = fmt.Sprintf("%s@%s", Name, Version)
	logFactory := avalogging.NewFactory(cfg.Logger)
	v.Log, err = logFactory.Make("ldvm-" + ctx.NodeID.String())
	if err != nil {
		return errp.Errorf("failed to create logger, %v", err)
	}

	logging.SetLogger(v.Log)
	v.Log.Info("LDVM.Initialize",
		zap.Uint32("networkID", ctx.NetworkID),
		zap.Stringer("subnetID", ctx.SubnetID),
		zap.Stringer("chainID", ctx.ChainID),
		zap.String("genesisData", string(genesisData)),
		zap.String("upgradeData", string(upgradeData)),
		zap.String("configData", string(configData)))

	err = v.initialize(cfg, genesisData, toEngine)
	if err != nil {
		v.Log.Error("LDVM.Initialize failed", zap.Error(err))
	}
	return errp.ErrorIf(err)
}

func (v *VM) initialize(
	cfg *config.Config,
	genesisData []byte,
	toEngine chan<- common.Message,
) (err error) {
	var gs *genesis.Genesis
	if len(genesisData) == 0 {
		genesisData = []byte(genesis.LocalGenesisConfigJSON)
	}

	gs, err = genesis.FromJSON(genesisData)
	if err != nil {
		return fmt.Errorf("parse genesis data error, %v", err)
	}

	// update the ChainID
	ld.SetChainID(gs.Chain.ChainID)

	chaindb := v.dbManager.Current().Database
	v.bc = chain.NewChain(v.ctx, cfg, gs, chaindb, toEngine, v.network.GossipTx)
	if err = v.bc.Bootstrap(); err != nil {
		return err
	}

	return nil
}

// SetState implements the common.VM SetState interface
// SetState communicates to VM its next state it starts
func (v *VM) SetState(state snow.State) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Log.Info("LDVM.SetState %v", zap.Stringer("state", state))
	return v.bc.SetState(state)
}

// Shutdown implements the common.VM Shutdown interface
// Shutdown is called when the node is shutting down.
func (v *VM) Shutdown() error {
	v.Log.Info("LDVM.Shutdown")
	// TODO graceful shutdown
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
		"rpc": {
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
	if err := server.RegisterService(api.NewBlockChainAPI(v.bc), Name); err != nil {
		v.Log.Error("LDVM.CreateHandlers error", zap.Error(err))
		return nil, err
	}

	ethAPI := api.NewEthAPI(v.bc, Version.String())
	cborAPI := api.NewAPI(v.bc, Version.String())
	v.Log.Info("CreateHandlers")
	return map[string]*common.HTTPHandler{
		"/rpc": {
			LockOptions: common.WriteLock,
			Handler:     server,
		},
		"/eth": {
			LockOptions: common.WriteLock,
			Handler:     ethAPI,
		},
		"/cborrpc": {
			LockOptions: common.WriteLock,
			Handler:     cborAPI,
		},
	}, nil
}

// HealthCheck implements the common.VM health.Checker HealthCheck interface
// Returns nil if the VM is healthy.
// Periodically called and reported via the node's Health API.
func (v *VM) HealthCheck() (interface{}, error) {
	return v.bc.HealthCheck()
}

// Connected implements the common.VM validators.Connector Connected interface
// Connector represents a handler that is called when a connection is marked as connected
func (v *VM) Connected(id ids.NodeID, nodeVersion *version.Application) error {
	v.Log.Info("LDVM.Connected",
		zap.Stringer("nodeID", id),
		zap.Stringer("version", nodeVersion))
	return nil // noop
}

// Disconnected implements the common.VM Disconnected interface
// Connector represents a handler that is called when a connection is marked as disconnected
func (v *VM) Disconnected(id ids.NodeID) error {
	v.Log.Info("LDVM.Disconnected", zap.Stringer("nodeID", id))
	return nil // noop
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
func (v *VM) AppRequest(ctx context.Context, id ids.NodeID, requestID uint32, deadline time.Time, request []byte) error {
	v.Log.Info("LDVM.AppRequest",
		zap.Stringer("nodeID", id),
		zap.Uint32("requestID", requestID),
		zap.Int("requestBytes", len(request)),
		zap.Time("deadline", deadline))
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
func (v *VM) AppResponse(ctx context.Context, id ids.NodeID, requestID uint32, response []byte) error {
	v.Log.Info("LDVM.AppResponse",
		zap.Stringer("nodeID", id),
		zap.Uint32("requestID", requestID),
		zap.Int("responseBytes", len(response)))
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
func (v *VM) AppRequestFailed(ctx context.Context, id ids.NodeID, requestID uint32) error {
	v.Log.Info("LDVM.AppRequestFailed",
		zap.Stringer("nodeID", id),
		zap.Uint32("requestID", requestID))
	return nil
}

// CrossChainAppRequest Notify this engine of a request for data from
// [chainID].
//
// The meaning of [request], and what should be sent in response to it, is
// application (VM) specific.
//
// Guarantees surrounding the request are specific to the implementation of
// the requesting VM. For example, the request may or may not be guaranteed
// to be well-formed/valid depending on the implementation of the requesting
// VM.
//
// This node should typically send a CrossChainAppResponse to [chainID] in
// response to a valid message using the same request ID before the
// deadline. However, the VM may arbitrarily choose to not send a response
// to this request.
func (v *VM) CrossChainAppRequest(ctx context.Context, chainID ids.ID, requestID uint32, deadline time.Time, request []byte) error {
	v.Log.Info("LDVM.CrossChainAppRequest",
		zap.Stringer("chainID", chainID),
		zap.Uint32("requestID", requestID),
		zap.Int("requestBytes", len(request)))
	return nil
}

// CrossChainAppRequestFailed notifies this engine that a
// CrossChainAppRequest message it sent to [chainID] with request ID
// [requestID] failed.
//
// This may be because the request timed out or because the message couldn't
// be sent to [chainID].
//
// It is guaranteed that:
// * This engine sent a request to [chainID] with ID [requestID].
// * CrossChainAppRequestFailed([chainID], [requestID]) has not already been
// called.
// * CrossChainAppResponse([chainID], [requestID]) has not already been
// called.
func (v *VM) CrossChainAppRequestFailed(ctx context.Context, chainID ids.ID, requestID uint32) error {
	v.Log.Info("LDVM.CrossChainAppRequestFailed",
		zap.Stringer("chainID", chainID),
		zap.Uint32("requestID", requestID))
	return nil
}

// CrossChainAppResponse notifies this engine of a response to the
// CrossChainAppRequest message it sent to [chainID] with request ID
// [requestID].
//
// The meaning of [response] is application (VM) specific.
//
// It is guaranteed that:
// * This engine sent a request to [chainID] with ID [requestID].
// * CrossChainAppRequestFailed([chainID], [requestID]) has not already been
// called.
// * CrossChainAppResponse([chainID], [requestID]) has not already been
// called.
//
// Guarantees surrounding the response are specific to the implementation of
// the responding VM. For example, the response may or may not be guaranteed
// to be well-formed/valid depending on the implementation of the requesting
// VM.
//
// If [response] is invalid or not the expected response, the VM chooses how
// to react. For example, the VM may send another CrossChainAppRequest, or
// it may give up trying to get the requested information.
func (v *VM) CrossChainAppResponse(ctx context.Context, chainID ids.ID, requestID uint32, response []byte) error {
	v.Log.Info("LDVM.CrossChainAppResponse",
		zap.Stringer("chainID", chainID),
		zap.Uint32("requestID", requestID),
		zap.Int("responseBytes", len(response)))
	return nil
}

// GetBlock implements the block.ChainVM GetBlock interface
//
// GetBlock attempt to fetch a block by it's ID
// If the block does not exist, an error should be returned.
func (v *VM) GetBlock(id ids.ID) (blk snowman.Block, err error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	blk, err = v.bc.GetBlock(util.Hash(id))
	if err != nil {
		v.Log.Error("LDVM.GetBlock", zap.Stringer("id", id), zap.Error(err))
	} else {
		v.Log.Info("LDVM.GetBlock",
			zap.Stringer("id", id),
			zap.Uint64("height", blk.Height()))
	}
	return
}

// ParseBlock implements the block.ChainVM ParseBlock interface
//
// ParseBlock attempt to fetch a block by its bytes.
func (v *VM) ParseBlock(data []byte) (blk snowman.Block, err error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	id := ids.ID(util.HashFromData(data))
	err = ld.Recover("", func() error {
		blk, err = v.bc.ParseBlock(data)
		return err
	})

	if err != nil {
		v.Log.Error("LDVM.ParseBlock", zap.Stringer("id", id), zap.Error(err))
	} else {
		v.Log.Info("LDVM.ParseBlock",
			zap.Stringer("id", id),
			zap.Uint64("height", blk.Height()))
	}
	return
}

// BuildBlock implements the block.ChainVM BuildBlock interface
//
// BuildBlock attempt to create a new block from data contained in the VM.
// If the VM doesn't want to issue a new block, an error should be returned.
func (v *VM) BuildBlock() (blk snowman.Block, err error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	blk, err = v.bc.BuildBlock()
	if err != nil {
		v.Log.Error("LDVM.BuildBlock", zap.Error(err))
	} else {
		v.Log.Info("LDVM.BuildBlock",
			zap.Stringer("id", blk.ID()),
			zap.Uint64("height", blk.Height()))
	}
	return
}

// SetPreference implements the block.ChainVM SetPreference interface
//
// SetPreference notify the VM of the currently preferred block.
// This should always be a block that has no children known to consensus.
func (v *VM) SetPreference(id ids.ID) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Log.Info("LDVM.SetPreference %s", zap.Stringer("id", id))
	err := v.bc.SetPreference(util.Hash(id))
	if err != nil {
		v.Log.Error("LDVM.SetPreference", zap.Stringer("id", id), zap.Error(err))
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
	blk := v.bc.LastAcceptedBlock()
	v.Log.Info("LDVM.LastAccepted",
		zap.Stringer("id", blk.ID()),
		zap.Uint64("height", blk.Height()))
	return blk.ID(), nil
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
	id, err := v.bc.GetBlockIDAtHeight(height)
	if err != nil {
		v.Log.Error("LDVM.GetBlockIDAtHeight %d error %v",
			zap.Uint64("height", height),
			zap.Error(err))
	} else {
		v.Log.Info("LDVM.GetBlockIDAtHeight",
			zap.Stringer("id", id),
			zap.Uint64("height", height))
	}
	return ids.ID(id), err
}
