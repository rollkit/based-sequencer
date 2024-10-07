package based_sequencer

import (
	"context"
	"encoding/binary"
	"github.com/dgraph-io/badger/v4"
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
	kv *badger.DB
}

var _ Store = &KVStore{}

func NewStore(underlying *badger.DB) *KVStore {
	return &KVStore{
		kv: underlying,
	}
}

func NewInMemoryStore() (*badger.DB, error) {
	opts := badger.DefaultOptions("").WithInMemory(true)
	return badger.Open(opts)
}

func (store *KVStore) Close() error {
	return store.kv.Close()
}

const (
	lastHeightKey    = "height"
	hashHashPrefix   = "hash/"
	hashHeightPrefix = "height/"
)

func (store *KVStore) GetLastDAHeight(ctx context.Context) (uint64, error) {
	var lastDAHeight uint64
	err := store.kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lastHeightKey))
		if err != nil {
			return err
		}
		bytes, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		lastDAHeight = binary.LittleEndian.Uint64(bytes)
		return nil
	})

	return lastDAHeight, err
}

func (store *KVStore) SetLastDAHeight(ctx context.Context, height uint64) error {
	return store.kv.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(lastHeightKey), binary.LittleEndian.AppendUint64(nil, height))
	})
}

func (store *KVStore) SetHashMapping(ctx context.Context, hash []byte, height uint64) error {
	return store.kv.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(hashHeightPrefix+string(hash)), binary.LittleEndian.AppendUint64(nil, height))
	})
}

func (store *KVStore) GetHashMapping(ctx context.Context, hash []byte) (uint64, error) {
	var hashMapping uint64
	err := store.kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hashHeightPrefix + string(hash)))
		if err != nil {
			return err
		}
		bytes, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		hashMapping = binary.LittleEndian.Uint64(bytes)
		return nil
	})
	return hashMapping, err
}

func (store *KVStore) SetNextHash(ctx context.Context, currentHash, nextHash []byte) error {
	return store.kv.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(hashHashPrefix+string(currentHash)), nextHash)
	})
}

func (store *KVStore) GetNextHash(ctx context.Context, currentHash []byte) ([]byte, error) {
	var nextHash []byte
	err := store.kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hashHashPrefix + string(currentHash)))
		if err != nil {
			return err
		}
		nextHash, err = item.ValueCopy(nil)
		return nil
	})
	return nextHash, err
}
