package based_sequencer

import (
	"context"
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
}

// NewSequencer initializes a BasedSequencer with the provided DA implementation.
func NewSequencer(da da.DA) *BasedSequencer {
	return &BasedSequencer{da: da}
}

var _ sequencing.Sequencer = (*BasedSequencer)(nil)

// SubmitRollupTransaction submits a transaction directly to DA, as a single blob.
func (b *BasedSequencer) SubmitRollupTransaction(ctx context.Context, rollupId sequencing.RollupId, tx sequencing.Tx) error {
	//TODO implement me
	panic("implement me")
}

// GetNextBatch reads data from namespace in DA and builds transactions batches.
func (b *BasedSequencer) GetNextBatch(ctx context.Context, lastBatchHash sequencing.Hash) (*sequencing.Batch, time.Time, error) {
	//TODO implement me
	panic("implement me")
}

// VerifyBatch ensures data-availability of a batch in DA.
func (b *BasedSequencer) VerifyBatch(ctx context.Context, batchHash sequencing.Hash) (bool, error) {
	//TODO implement me
	panic("implement me")
}
