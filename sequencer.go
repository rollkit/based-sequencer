package based_sequencer

import (
	"context"
	"time"

	"github.com/rollkit/go-da"
	"github.com/rollkit/go-sequencing"
)

type BasedSequencer struct {
	da da.DA
}

var _ sequencing.Sequencer = (*BasedSequencer)(nil)

func (b *BasedSequencer) SubmitRollupTransaction(ctx context.Context, rollupId sequencing.RollupId, tx sequencing.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (b *BasedSequencer) GetNextBatch(ctx context.Context, lastBatchHash sequencing.Hash) (*sequencing.Batch, time.Time, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BasedSequencer) VerifyBatch(ctx context.Context, batchHash sequencing.Hash) (bool, error) {
	//TODO implement me
	panic("implement me")
}
