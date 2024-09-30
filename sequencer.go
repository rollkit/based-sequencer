package based_sequencer

import (
	"context"
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
}

// NewSequencer initializes a BasedSequencer with the provided DA implementation.
func NewSequencer(da da.DA) *BasedSequencer {
	seq := &BasedSequencer{da: da}
	seq.daHeight.Store(1) // TODO(tzdybal): this needs to be config or constructor param
	return seq
}

var _ sequencing.Sequencer = (*BasedSequencer)(nil)

// SubmitRollupTransaction submits a transaction directly to DA, as a single blob.
func (b *BasedSequencer) SubmitRollupTransaction(ctx context.Context, rollupId sequencing.RollupId, tx sequencing.Tx) error {
	_, err := b.da.Submit(ctx, []da.Blob{tx}, 0, rollupId)
	if err != nil {
		return err
	}
	return nil
}

// GetNextBatch reads data from namespace in DA and builds transactions batches.
func (b *BasedSequencer) GetNextBatch(ctx context.Context, lastBatchHash sequencing.Hash) (*sequencing.Batch, time.Time, error) {
	// TODO(tzdybal): this needs to be a field
	namespace := []byte("test namespace")

	var err error
	result := &da.GetIDsResult{}

	for ctx.Err() == nil && len(result.IDs) == 0 {
		result, err = b.da.GetIDs(ctx, b.daHeight.Load(), namespace)
		if err != nil {
			if strings.Contains(err.Error(), "given height is from the future") {
				time.Sleep(100 * time.Millisecond) // TODO(tzdybal): this needs to be configurable
				continue
			}
			return nil, time.Time{}, err
		}
		b.daHeight.Add(1)
	}

	blobs, err := b.da.Get(ctx, result.IDs, namespace)
	if err != nil {
		return nil, time.Time{}, err
	}

	return &sequencing.Batch{Transactions: blobs}, result.Timestamp, nil
}

// VerifyBatch ensures data-availability of a batch in DA.
func (b *BasedSequencer) VerifyBatch(ctx context.Context, batchHash sequencing.Hash) (bool, error) {
	//TODO implement me
	panic("implement me")
}
