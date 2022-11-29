// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txpool

import (
	"context"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/ids"
)

const MaxObjectSize = 1 << 20 // 1MB

const AcquireRemainingLife = int64(time.Minute * 10)

const (
	TxsBucket   = "txs"
	BatchBucket = "txs:batch"
)

// Object represents a data with LDC's block height in the POS.
type Object struct {
	Raw []byte
	// -2: rejected, will be remove at expiration time
	// -1: wait for build, will be remove if not accepted before expiration time
	// 0: processing, will be remove if not accepted before expiration time, except genesis tx objects;
	// > 0: accepted, the height of block, the object is permanently stored.
	Height int64
}

// Hash returns the SHA3-256 digest of the object's data.
// It is the key of the object.
func (obj *Object) Hash() ids.ID32 {
	return ids.ID32FromData(obj.Raw)
}

type Objects []*Object

func (os Objects) ToRawList() []cbor.RawMessage {
	rawList := make([]cbor.RawMessage, 0, len(os))
	for _, obj := range os {
		rawList = append(rawList, obj.Raw)
	}
	return rawList
}

// POS is permanent object storage interface.
type POS interface {
	GetObject(ctx context.Context, bucket string, hash ids.ID32) (*Object, error)
	PutObject(ctx context.Context, bucket string, objectRaw []byte) error
	RemoveObject(ctx context.Context, bucket string, hash ids.ID32) error
	BatchAcquire(ctx context.Context, bucket string, hashList ids.IDList[ids.ID32]) error
	BatchAccept(ctx context.Context, bucket string, height uint64, hashList ids.IDList[ids.ID32]) error
	ListUnaccept(ctx context.Context, bucket, token string) (hashList ids.IDList[ids.ID32], nextToken string, err error)
}
