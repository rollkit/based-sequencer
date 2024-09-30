package based_sequencer

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	ds "github.com/ipfs/go-datastore"

	"github.com/rollkit/go-da"
	"github.com/rollkit/go-sequencing"
)

// BasedSequencer implements go-sequencing API with based sequencing logic.
//
// All transactions are passed directly to DA to be saved in a namespace. Each transaction is submitted as separate blob.
// When batch is requested DA blocks are scanned to read all blobs from given namespace at given height.
type BasedSequencer struct {
	da da.DA

	daHeight atomic.Uint64

	store Store
}

// NewSequencer initializes a BasedSequencer with the provided DA implementation.
func NewSequencer(da da.DA) *BasedSequencer {
	seq := &BasedSequencer{da: da}
	seq.daHeight.Store(1)  // TODO(tzdybal): this needs to be config or constructor param
	seq.store = &KVStore{} // TODO(tzdybal): extract parameter
	return seq
}

var _ sequencing.Sequencer = (*BasedSequencer)(nil)

// SubmitRollupTransaction submits a transaction directly to DA, as a single blob.
func (seq *BasedSequencer) SubmitRollupTransaction(ctx context.Context, rollupId sequencing.RollupId, tx sequencing.Tx) error {
	_, err := seq.da.Submit(ctx, []da.Blob{tx}, 0, rollupId)
	if err != nil {
		return err
	}
	return nil
}

// GetNextBatch reads data from namespace in DA and builds transactions batches.
func (seq *BasedSequencer) GetNextBatch(ctx context.Context, lastBatchHash sequencing.Hash) (*sequencing.Batch, time.Time, error) {
	// optimistic case - we already know the hash after lastBatchHash
	nextBatchHash, err := seq.store.GetNextHash(ctx, lastBatchHash)
	if err == nil {
		height, err := seq.store.GetHashMapping(ctx, nextBatchHash)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("store contents inconsistent: %w", err)
		}
		return seq.getBatchByHeight(ctx, height)
	}

	// if there is no indexing information in store, it's not an error
	if !errors.Is(err, ds.ErrNotFound) {
		return nil, time.Time{}, err
	}

	// we need to search for next batch
	for ctx.Err() == nil {
		batch, ts, err := seq.getBatchByHeight(ctx, seq.daHeight.Load())
		if err != nil {
			if strings.Contains(err.Error(), "given height is from the future") {
				time.Sleep(100 * time.Millisecond) // TODO(tzdybal): this needs to be configurable
				continue
			}
			return nil, time.Time{}, err
		}

		// TODO(tzdybal): extract method
		hash, err := BatchHash(batch)
		if err != nil {
			return nil, time.Time{}, err
		}
		if err = seq.store.SetNextHash(ctx, lastBatchHash, hash); err != nil {
			return nil, time.Time{}, err
		}
		if err = seq.store.SetHashMapping(ctx, hash, seq.daHeight.Load()); err != nil {
			return nil, time.Time{}, err
		}
		seq.daHeight.Add(1)
		return batch, ts, err
	}

	return nil, time.Time{}, ctx.Err()
}

// VerifyBatch ensures data-availability of a batch in DA.
func (seq *BasedSequencer) VerifyBatch(ctx context.Context, batchHash sequencing.Hash) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func BatchHash(batch *sequencing.Batch) (sequencing.Hash, error) {
	batchBytes, err := batch.Marshal()
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(batchBytes)
	return h[:], nil
}

func (seq *BasedSequencer) getBatchByHeight(ctx context.Context, height uint64) (*sequencing.Batch, time.Time, error) {
	// TODO(tzdybal): this needs to be a field
	namespace := []byte("test namespace")
	result, err := seq.da.GetIDs(ctx, height, namespace)
	if err != nil {
		return nil, time.Time{}, err
	}

	blobs, err := seq.da.Get(ctx, result.IDs, namespace)
	if err != nil {
		return nil, time.Time{}, err
	}

	return &sequencing.Batch{Transactions: blobs}, result.Timestamp, nil
}
