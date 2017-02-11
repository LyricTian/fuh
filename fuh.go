package fuh

import (
	"io"
	"net/http"
)

var (
	uh = &uploadHandle{}
)

// SetConfig set upload the configuration parameters
func SetConfig(cfg *UploadConfig) {
	uh.setConfig(cfg)
}

// SetStore set upload data storage interface
func SetStore(store Storer) {
	uh.setStore(store)
}

// Upload upload a single file
func Upload(r *http.Request, key string, fnh FileNameHandle, fsh FileSizeHandle) (*FileInfo, error) {
	return uh.Upload(r, key, fnh, fsh)
}

// UploadMulti upload multiple files
func UploadMulti(r *http.Request, key string, fnh FileNameHandle, fsh FileSizeHandle) ([]*FileInfo, error) {
	return uh.UploadMulti(r, key, fnh, fsh)
}

// UploadReader upload data stream file
func UploadReader(r io.Reader, fnh FileNameHandle, fsh FileSizeHandle) (*FileInfo, error) {
	return uh.UploadReader(r, fnh, fsh)
}
