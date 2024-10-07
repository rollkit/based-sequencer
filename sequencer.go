package based_sequencer

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"strings"
	"sync/atomic"
	"time"

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
func NewSequencer(da da.DA) (*BasedSequencer, error) {
	seq := &BasedSequencer{da: da}
	seq.daHeight.Store(1) // TODO(tzdybal): this needs to be config or constructor param
	inMem, err := NewInMemoryStore()
	if err != nil {
		return nil, err
	}
	seq.store = NewStore(inMem) // TODO(tzdybal): extract parameter
	return seq, nil
}

var _ sequencing.Sequencer = (*BasedSequencer)(nil)

// SubmitRollupTransaction submits a transaction directly to DA, as a single blob.
func (seq *BasedSequencer) SubmitRollupTransaction(ctx context.Context, req sequencing.SubmitRollupTransactionRequest) (*sequencing.SubmitRollupTransactionResponse, error) {
	_, err := seq.da.Submit(ctx, []da.Blob{req.Tx}, 0, req.RollupId)
	if err != nil {
		return nil, err
	}
	return &sequencing.SubmitRollupTransactionResponse{}, nil
}

// GetNextBatch reads data from namespace in DA and builds transactions batches.
func (seq *BasedSequencer) GetNextBatch(ctx context.Context, req sequencing.GetNextBatchRequest) (*sequencing.GetNextBatchResponse, error) {
	// optimistic case - we already know the hash after lastBatchHash
	nextBatchHash, err := seq.store.GetNextHash(ctx, req.LastBatchHash)
	if err == nil {
		height, err := seq.store.GetHashMapping(ctx, nextBatchHash)
		if err != nil {
			return nil, fmt.Errorf("kv contents inconsistent: %w", err)
		}
		return seq.getBatchByHeight(ctx, height)
	}

	// if there is no indexing information in kv, it's not an error
	if !errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	}

	// we need to search for next batch
	for ctx.Err() == nil {
		resp, err := seq.getBatchByHeight(ctx, seq.daHeight.Load())
		if err != nil {
			if strings.Contains(err.Error(), "given height is from the future") {
				time.Sleep(100 * time.Millisecond) // TODO(tzdybal): this needs to be configurable
				continue
			}
			return nil, err
		}

		// TODO(tzdybal): extract method
		hash, err := resp.Batch.Hash()
		if err != nil {
			return nil, err
		}
		if err = seq.store.SetNextHash(ctx, req.LastBatchHash, hash); err != nil {
			return nil, err
		}
		if err = seq.store.SetHashMapping(ctx, hash, seq.daHeight.Load()); err != nil {
			return nil, err
		}
		seq.daHeight.Add(1)
		return resp, err
	}

	return nil, ctx.Err()
}

// VerifyBatch ensures data-availability of a batch in DA.
func (seq *BasedSequencer) VerifyBatch(ctx context.Context, req sequencing.VerifyBatchRequest) (*sequencing.VerifyBatchResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (seq *BasedSequencer) getBatchByHeight(ctx context.Context, height uint64) (*sequencing.GetNextBatchResponse, error) {
	// TODO(tzdybal): this needs to be a field
	namespace := []byte("test namespace")
	result, err := seq.da.GetIDs(ctx, height, namespace)
	if err != nil {
		return nil, err
	}

	blobs, err := seq.da.Get(ctx, result.IDs, namespace)
	if err != nil {
		return nil, err
	}

	return &sequencing.GetNextBatchResponse{Batch: &sequencing.Batch{Transactions: blobs}, Timestamp: result.Timestamp}, nil
}
