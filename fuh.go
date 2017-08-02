package fuh

import (
	"context"
	"io"
	"net/http"
)

// FileInfo upload the basic information of the file
type FileInfo interface {
	FullName() string
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
	upload *uploadHandle
)

// SetConfig set the configuration parameters
func SetConfig(cfg *Config) {
	newUploader().cfg = cfg
}

// SetStore set storage
func SetStore(store Storer) {
	newUploader().store = store
}

// Upload file upload
func Upload(ctx context.Context, r *http.Request, key string) ([]FileInfo, error) {
	return newUploader().Upload(ctx, r, key)
}

func newUploader() *uploadHandle {
	if upload == nil {
		upload = &uploadHandle{}
	}
	return upload
}
