package based_sequencer

import "context"

type Store interface {
	GetLastDAHeight(ctx context.Context) (uint64, error)
	SetLastDAHeight(ctx context.Context, height uint64) error

	SetHashMapping(ctx context.Context, hash []byte, height uint64) error
	GetHashMapping(ctx context.Context, hash []byte) (uint64, error)

	SetNextHash(ctx context.Context, currentHash, nextHash []byte) error
	GetNextHash(ctx context.Context, currentHash []byte) ([]byte, error)
}

type KVStore struct {
}

var _ Store = &KVStore{}

func (K *KVStore) GetLastDAHeight(ctx context.Context) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (K *KVStore) SetLastDAHeight(ctx context.Context, height uint64) error {
	//TODO implement me
	panic("implement me")
}

func (K *KVStore) SetHashMapping(ctx context.Context, hash []byte, height uint64) error {
	//TODO implement me
	panic("implement me")
}

func (K *KVStore) GetHashMapping(ctx context.Context, hash []byte) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (K *KVStore) SetNextHash(ctx context.Context, currentHash, nextHash []byte) error {
	//TODO implement me
	panic("implement me")
}

func (K *KVStore) GetNextHash(ctx context.Context, currentHash []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
