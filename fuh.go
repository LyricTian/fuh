package fuh

import (
	"context"
	"io"
	"net/http"
	"sync"
)

// FileInfo upload the basic information of the file
type FileInfo interface {
	FullName() string
	Name() string
	Size() int64
}

// Uploader file upload interface
type Uploader interface {
	Upload(ctx context.Context, r *http.Request, key string) ([]FileInfo, error)
}

// Storer file storage interface
type Storer interface {
	Store(ctx context.Context, filename string, data io.Reader, size int64) error
}

var (
	internalHandle *uploadHandle
	once           sync.Once
)

func uploader() *uploadHandle {
	once.Do(func() {
		internalHandle = &uploadHandle{}
	})
	return internalHandle
}

// SetConfig set the configuration parameters
func SetConfig(cfg *Config) {
	uploader().cfg = cfg
}

// SetStore set storage
func SetStore(store Storer) {
	uploader().store = store
}

// Upload file upload
func Upload(ctx context.Context, r *http.Request, key string) ([]FileInfo, error) {
	return uploader().Upload(ctx, r, key)
}
