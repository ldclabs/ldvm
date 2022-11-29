// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txpool

import (
	"context"
	"fmt"

	"github.com/mailgun/holster/v4/collections"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/sync"
)

const (
	knownTxsCapacity = 1000000
	knownTxsTTL      = 600 // seconds
)

// TxPool contains all currently known transactions.
type TxPool struct {
	mu             sync.RWMutex
	builderKeepers signer.Keys
	builder        string
	cwtAudience    string
	cwtExData      []byte
	pendingIDs     ids.Set[string]
	pendingTxs     ld.Txs
	pendingBatch   map[ids.ID32]ids.IDList[ids.ID32]
	knownTxs       *knownTxObjects
	pos            POS
	chain          sync.Value[Chain]
	opts           TxPoolOptions
}

type TxPoolOptions struct {
	Builder     ids.Address
	CWTAudience string
	CWTExData   []byte
}

// NewTxPool creates a new transaction pool.
func NewTxPool(pos POS, chain Chain, opts TxPoolOptions) *TxPool {
	pool := &TxPool{
		builder:      opts.Builder.String(),
		cwtAudience:  opts.CWTAudience,
		cwtExData:    opts.CWTExData,
		pendingIDs:   ids.NewSet[string](10000),
		pendingTxs:   make([]*ld.Transaction, 0, 10000),
		pendingBatch: make(map[ids.ID32]ids.IDList[ids.ID32]),
		knownTxs:     &knownTxObjects{collections.NewTTLMap(knownTxsCapacity)},
		pos:          pos,
		opts:         opts,
	}
	pool.chain.Store(chain)
	return pool
}

type TxOrBatch struct {
	Tx           *ld.TxData  `cbor:"tx,omitempty"`
	Signatures   signer.Sigs `cbor:"ss,omitempty"`
	ExSignatures signer.Sigs `cbor:"es,omitempty"`
	Batch        ld.Txs      `cbor:"ba,omitempty"`
}

func (t *TxOrBatch) ToTransaction() (*ld.Transaction, error) {
	switch {
	case t.Tx != nil:
		tx := &ld.Transaction{Tx: *t.Tx, Signatures: t.Signatures, ExSignatures: t.ExSignatures}
		if err := tx.SyntacticVerify(); err != nil {
			return nil, err
		}
		return tx, nil

	case len(t.Batch) > 1:
		return ld.NewBatchTx(t.Batch...)

	default:
		return nil, fmt.Errorf("invalid TxOrBatch")
	}
}

type TxsBuildStatus struct {
	Unknown  ids.IDList[ids.ID32] `cbor:"u,omitempty"`
	Rejected ids.IDList[ids.ID32] `cbor:"r,omitempty"`
}

func (p *TxPool) Bootstrap(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check&Init Genesis Txs
	{
		txs, err := p.chain.MustLoad().GetGenesisTxs(ctx)
		if err != nil {
			return err
		}

		for i := range txs {
			tx := txs[i]
			p.knownTxs.set(string(tx.ID[:]), &Object{Raw: tx.Bytes(), Height: 0})
		}

		if obj, err := p.loadByID(ctx, txs[len(txs)-1].ID); err != nil || obj.Height < 0 {
			for i := range txs {
				_ = p.pos.PutObject(ctx, TxsBucket, txs[i].Bytes()) // ignore error
			}

			if err = p.pos.BatchAccept(ctx, TxsBucket, 0, txs.IDs()); err != nil {
				return nil
			}
		}
	}

	// Load builder's account
	{
		builderAcc, err := p.chain.MustLoad().GetAccount(ctx, p.opts.Builder)
		if err != nil {
			return err
		}
		p.builderKeepers = builderAcc.Keepers.Clone()
	}

	// Load txs by batch from POS
	{
		var err error
		var nextToken string
		var list ids.IDList[ids.ID32]
		var batchs ids.IDList[ids.ID32]

		for {
			list, nextToken, err = p.pos.ListUnaccept(ctx, BatchBucket, nextToken)
			if err != nil {
				return err
			}
			batchs = append(batchs, list...)
			if err = batchs.CheckDuplicate(); err != nil {
				return err
			}

			if nextToken == "" || len(list) == 0 {
				break
			}
		}

		for _, bid := range batchs {
			obj, err := p.pos.GetObject(ctx, BatchBucket, bid)
			if err != nil {
				return err
			}
			if obj.Hash() != bid {
				return fmt.Errorf("invalid batch %s", bid.String())
			}

			var txIDs ids.IDList[ids.ID32]
			if err = encoding.UnmarshalCBOR(obj.Raw, &txIDs); err != nil {
				return err
			}

			if len(txIDs) <= 1 {
				return fmt.Errorf("invalid batch %s", bid.String())
			}

			objs, err := p.loadByIDs(ctx, txIDs)
			if err != nil {
				return err
			}
			txs, err := TxsFrom(objs, false)
			if err != nil {
				return err
			}

			tx, err := ld.NewBatchTx(txs...)
			if err != nil {
				return err
			}

			p.pendingBatch[bid] = txIDs
			for i, id := range txIDs {
				sid := string(id[:])
				p.pendingIDs.Add(sid)
				p.knownTxs.set(sid, objs[i])
			}
			p.pendingTxs = append(p.pendingTxs, tx)
		}
	}

	// Load txs out of batch from POS
	{
		var err error
		var nextToken string
		var list ids.IDList[ids.ID32]
		var txIDs ids.IDList[ids.ID32]

		for {
			list, nextToken, err = p.pos.ListUnaccept(ctx, TxsBucket, nextToken)
			if err != nil {
				return err
			}
			txIDs = append(txIDs, list...)
			if err = txIDs.CheckDuplicate(); err != nil {
				return err
			}

			if nextToken == "" || len(list) == 0 {
				break
			}
		}

		for _, id := range txIDs {
			sid := string(id[:])
			if p.pendingIDs.Has(sid) {
				continue
			}
			obj, err := p.loadByID(ctx, id)
			if err != nil {
				return err
			}
			tx, err := TxFrom(obj, true)
			if err != nil {
				return err
			}
			p.pendingIDs.Add(sid)
			p.knownTxs.set(sid, obj)
			p.pendingTxs = append(p.pendingTxs, tx)
		}
	}

	return nil
}

func (p *TxPool) SetChain(chain Chain) {
	p.chain.Store(chain)
}

func (p *TxPool) rpcSizeToBuild(ctx context.Context) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.pendingIDs.Len()
}

func (p *TxPool) rpcSubmitTxs(ctx context.Context, txs ld.Txs) error {
	var err error
	var tx *ld.Transaction

	switch len(txs) {
	case 0:
		return fmt.Errorf("no txs")
	case 1:
		tx = txs[0]
		if err = tx.SyntacticVerify(); err != nil {
			return err
		}
	default:
		if tx, err = ld.NewBatchTx(txs...); err != nil {
			return err
		}
	}

	for i := range txs {
		if l := len(txs[i].Bytes()); l > MaxObjectSize {
			return fmt.Errorf("tx %s too large, expected <= %d, got %d", txs[i].ID.String(), MaxObjectSize, l)
		}
	}

	if err := p.chain.MustLoad().PreVerifyTxs(ctx, txs); err != nil {
		return err
	}

	for i := range txs {
		if err := p.pos.PutObject(ctx, TxsBucket, txs[i].Bytes()); err != nil {
			return err
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	switch {
	case tx.IsBatched():
		txIDs := tx.IDs()
		bid := ids.ID32FromData(encoding.MustMarshalCBOR(txIDs))
		p.pendingBatch[bid] = txIDs
		p.pendingTxs = append(p.pendingTxs, tx)
		for i := range txs {
			sid := string(txs[i].ID[:])
			p.pendingIDs.Add(sid)
			p.knownTxs.set(sid, &Object{Raw: txs[i].Bytes(), Height: -1})
		}

	default:
		sid := string(tx.ID[:])
		p.pendingTxs = append(p.pendingTxs, tx)
		p.pendingIDs.Add(sid)
		p.knownTxs.set(sid, &Object{Raw: tx.Bytes(), Height: -1})
	}

	return nil
}

func (p *TxPool) rpcLoadByIDs(ctx context.Context, txIDs ids.IDList[ids.ID32]) (Objects, error) {
	objs := make(Objects, 0, len(txIDs))

	var err error
	for _, id := range txIDs {
		sid := string(id[:])
		obj := p.knownTxs.get(sid)
		if obj != nil {
			objs = append(objs, obj)
			continue
		}

		obj, err = p.loadByID(ctx, id)
		if err != nil {
			return nil, err
		}
		p.knownTxs.set(sid, obj)
		objs = append(objs, obj)
	}
	return objs, nil
}

func (p *TxPool) rpcFetchToBuild(ctx context.Context, amount int) ([]TxOrBatch, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pendingTxs.Sort()
	if amount > len(p.pendingTxs) {
		amount = len(p.pendingTxs)
	}

	n := 0
	offset := 0
	result := make([]TxOrBatch, 0, amount)
	txIDs := make(ids.IDList[ids.ID32], 0, amount)
	for i := range p.pendingTxs {
		tx := p.pendingTxs[i]
		n += tx.Size()
		if n > amount {
			break
		}

		offset += 1
		switch {
		case tx.IsBatched():
			result = append(result, TxOrBatch{Batch: tx.Txs()})
		default:
			result = append(result, TxOrBatch{
				Tx:           &tx.Tx,
				Signatures:   tx.Signatures,
				ExSignatures: tx.ExSignatures,
			})
		}

		for _, id := range tx.IDs() {
			sid := string(id[:])
			p.pendingIDs.Del(sid)
			p.knownTxs.setHeight(sid, 0)
			txIDs = append(txIDs, id)
		}
	}

	if offset > 0 {
		n = copy(p.pendingTxs, p.pendingTxs[offset:])
		p.pendingTxs = p.pendingTxs[:n]
	}

	if err := p.pos.BatchAcquire(ctx, TxsBucket, txIDs); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *TxPool) rpcUpdateBuildStatus(ctx context.Context, ts *TxsBuildStatus) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, id := range ts.Rejected {
		p.knownTxs.setHeight(string(id[:]), -2)
	}

	bids := p.findBatchs(ts.Rejected)
	for _, bid := range bids {
		delete(p.pendingBatch, bid)
		go p.pos.RemoveObject(ctx, BatchBucket, bid)
	}

	for _, id := range ts.Unknown {
		p.knownTxs.setHeight(string(id[:]), -1)
	}
	bids = p.findBatchs(ts.Unknown)
	for _, bid := range bids {
		txIDs := p.pendingBatch[bid]
		objs, err := p.rpcLoadByIDs(ctx, txIDs)
		if err != nil {
			continue
		}
		txs, err := TxsFrom(objs, false)
		if err != nil {
			continue
		}

		tx, err := ld.NewBatchTx(txs...)
		if err != nil {
			continue
		}

		for i, id := range txIDs {
			sid := string(id[:])
			p.pendingIDs.Add(sid)
			p.knownTxs.set(sid, objs[i])
		}
		p.pendingTxs = append(p.pendingTxs, tx)
	}

	return nil
}

func (p *TxPool) rpcAcceptByBlock(ctx context.Context, blk *ld.Block) error {
	if err := p.pos.BatchAccept(ctx, TxsBucket, blk.Height, blk.Txs); err != nil {
		return err
	}

	for _, id := range blk.Txs {
		p.knownTxs.setHeight(string(id[:]), int64(blk.Height))
	}

	bids := p.findBatchs(blk.Txs)
	for _, bid := range bids {
		delete(p.pendingBatch, bid)
		go p.pos.RemoveObject(ctx, BatchBucket, bid)
	}

	return nil
}

func (p *TxPool) rpcUpdateBuilderKeepers(ctx context.Context) error {
	builderAcc, err := p.chain.MustLoad().GetAccount(ctx, p.opts.Builder)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.builderKeepers = builderAcc.Keepers.Clone()
	return nil
}

func (p *TxPool) findBatchs(txIDs ids.IDList[ids.ID32]) ids.IDList[ids.ID32] {
	batchs := make(ids.IDList[ids.ID32], 0)
	for i, id := range txIDs {
		for bid, batch := range p.pendingBatch {
			if batch.Has(id) {
				batchs = append(batchs, bid)
				i += len(batch)
				break
			}
		}
	}
	return batchs
}

func (p *TxPool) loadByIDs(ctx context.Context, txIDs ids.IDList[ids.ID32]) (Objects, error) {
	objects := make(Objects, 0, len(txIDs))
	for _, id := range txIDs {
		obj, err := p.loadByID(ctx, id)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func (p *TxPool) loadByID(ctx context.Context, txID ids.ID32) (*Object, error) {
	obj, err := p.pos.GetObject(ctx, TxsBucket, txID)
	if err != nil {
		return nil, err
	}
	if obj.Hash() != txID {
		return nil, fmt.Errorf("invalid tx %s", txID.String())
	}

	return obj, nil
}

type knownTxObjects struct {
	cache *collections.TTLMap
}

func (k *knownTxObjects) get(txID string) *Object {
	if s, ok := k.cache.Get(txID); ok {
		return s.(*Object)
	}
	return nil
}

func (k *knownTxObjects) set(txID string, obj *Object) {
	if obj != nil {
		k.cache.Set(txID, obj, knownTxsTTL)
	}
}

func (k *knownTxObjects) setHeight(txID string, height int64) {
	if obj := k.get(txID); obj != nil {
		obj.Height = height
		if height < -1 { // rejected tx object
			obj.Raw = nil
		}
	}
}

func TxFrom(obj *Object, verify bool) (*ld.Transaction, error) {
	tx := &ld.Transaction{}
	if err := tx.Unmarshal(obj.Raw); err != nil {
		return nil, err
	}
	if verify {
		if err := tx.SyntacticVerify(); err != nil {
			return nil, err
		}
	}
	if obj.Height > -1 {
		tx.Height = uint64(obj.Height)
	}
	return tx, nil
}

func TxsFrom(os Objects, verify bool) (ld.Txs, error) {
	txs := make(ld.Txs, 0, len(os))
	for _, obj := range os {
		tx, err := TxFrom(obj, verify)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
