package based_sequencer

import (
	"context"
	"encoding/binary"
	"time"

	ds "github.com/ipfs/go-datastore"
	badger4 "github.com/ipfs/go-ds-badger4"
)

type Store interface {
	GetLastDAHeight(ctx context.Context) (uint64, error)
	SetLastDAHeight(ctx context.Context, height uint64) error

	SetHashMapping(ctx context.Context, hash []byte, height uint64) error
	GetHashMapping(ctx context.Context, hash []byte) (uint64, error)

	SetNextHash(ctx context.Context, currentHash, nextHash []byte) error
	GetNextHash(ctx context.Context, currentHash []byte) ([]byte, error)
}

type KVStore struct {
	store ds.Batching
}

var _ Store = &KVStore{}

func NewStore(underlying ds.Batching) *KVStore {
	return &KVStore{
		store: underlying,
	}
}
func NewInMemoryStore() (ds.Batching, error) {
	inMemoryOptions := &badger4.Options{
		GcDiscardRatio: 0.2,
		GcInterval:     15 * time.Minute,
		GcSleep:        10 * time.Second,
		Options:        badger4.DefaultOptions.WithInMemory(true),
	}
	return badger4.NewDatastore("", inMemoryOptions)

}

const (
	lastHeightKey    = "height"
	hashHashPrefix   = "hash/"
	hashHeightPrefix = "height/"
)

func (store *KVStore) GetLastDAHeight(ctx context.Context) (uint64, error) {
	bytes, err := store.store.Get(ctx, ds.NewKey(lastHeightKey))
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(bytes), nil
}

func (store *KVStore) SetLastDAHeight(ctx context.Context, height uint64) error {
	return store.store.Put(ctx, ds.NewKey(lastHeightKey), binary.LittleEndian.AppendUint64(nil, height))
}

func (store *KVStore) SetHashMapping(ctx context.Context, hash []byte, height uint64) error {
	return store.store.Put(ctx, ds.NewKey(hashHeightPrefix+string(hash)), binary.LittleEndian.AppendUint64(nil, height))
}

func (store *KVStore) GetHashMapping(ctx context.Context, hash []byte) (uint64, error) {
	bytes, err := store.store.Get(ctx, ds.NewKey(hashHeightPrefix+string(hash)))
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(bytes), nil
}

func (store *KVStore) SetNextHash(ctx context.Context, currentHash, nextHash []byte) error {
	return store.store.Put(ctx, ds.NewKey(hashHashPrefix+string(currentHash)), nextHash)
}

func (store *KVStore) GetNextHash(ctx context.Context, currentHash []byte) ([]byte, error) {
	nextHash, err := store.store.Get(ctx, ds.NewKey(hashHashPrefix+string(currentHash)))
	if err != nil {
		return nil, err
	}
	return nextHash, nil
}
