package storage

import "errors"

var (
	ErrNotFound = errors.New("not found.")
)

type MetadataStorager interface {
	Get(key string, v interface{}) error
	Put(key string, v interface{}) error
	Delete(key string) error
}

type DataStorager interface {
	Bytes(key string) ([]byte, error)
	PutBytes(key string, val []byte) error
	Delete(key string) error
}

type SimpleStorager interface {
	Bytes(key string) ([]byte, error)
	PutBytes(key string, val []byte) error
	Delete(key string) error
}
