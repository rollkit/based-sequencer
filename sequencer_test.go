package based_sequencer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollkit/go-da"
	"github.com/rollkit/go-da/mocks"
	"github.com/rollkit/go-sequencing"
)

func TestNewSequencer(t *testing.T) {
	mockDA := mocks.NewMockDA(t)
	seq := NewSequencer(mockDA)
	require.NotNil(t, seq)
}

// TestSubmitRollupTransaction ensures that single rollup transaction submitted to sequencer is submitted as one blob to DA.
func TestSubmitRollupTransaction(t *testing.T) {
	mockDA := mocks.NewMockDA(t)
	seq := NewSequencer(mockDA)
	require.NotNil(t, seq)

	const (
		rollupId = "test rollup"
		testTx   = "this is a random transaction"
	)

	// make sure that sequencer submits only one blob, and transaction is included in this blob
	// this is not `Equals` test, because the actual tx -> blob mapping is not yet defined
	mockDA.On("Submit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		blobs, ok := args.Get(1).([]da.Blob)
		require.True(t, ok)
		require.Len(t, blobs, 1)
		require.Contains(t, string(blobs[0]), testTx)
	}).Return(nil, nil).Once()

	err := seq.SubmitRollupTransaction(context.Background(), sequencing.RollupId(rollupId), sequencing.Tx(testTx))
	require.NoError(t, err)

	mockDA.AssertExpectations(t)
}

func TestGetNextBatch(t *testing.T) {
	var (
		testNamespace = []byte("test namespace")
		//blockTime     = 1 * time.Second
	)

	mockDA := mocks.NewMockDA(t)

	// sequencer needs to ask about MaxBytes
	mockDA.On("MaxBlobSize", mock.Anything).Return(int64(1000), nil)
	// some IDs are returned, so mocked blobs can be easily handled
	mockDA.On("GetIDs", mock.Anything, uint64(1), testNamespace).Return(&da.GetIDsResult{
		IDs:       []da.ID{{1}, {2}, {3}},
		Timestamp: time.Now(),
	}, nil).Once()
	//mockDA.On("GetIDs", mock.Anything, uint64(2), testNamespace).Return(da.GetIDsResult{
	//	IDs:       []da.ID{{4}, {4}},
	//	Timestamp: time.Now().Add(blockTime),
	//}, nil).Once()
	//mockDA.On("GetIDs", mock.Anything, uint64(3), testNamespace).Return(da.GetIDsResult{
	//	IDs:       []da.ID{{6}, {7}, {8}, {9}},
	//	Timestamp: time.Now().Add(2 * blockTime),
	//}, nil).Once()

	transactions := make([][]byte, 10)
	for i := 0; i < len(transactions); i++ {
		transactions[i] = []byte(fmt.Sprintf("transaction %d", i))
	}
	mockDA.On("Get", mock.Anything, mock.Anything, testNamespace).Return(
		func(ctx context.Context, ids []da.ID, namespace da.Namespace) ([]da.Blob, error) {
			blobs := make([]da.Blob, len(ids))
			for i := range ids {
				blobs[i] = transactions[i]
			}
			return blobs, nil
		})

	seq := NewSequencer(mockDA)
	require.NotNil(t, seq)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)

	batch, ts, err := seq.GetNextBatch(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, batch)
	require.NotEmpty(t, ts)

	mockDA.AssertExpectations(t)
}
