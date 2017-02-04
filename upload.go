package fuh

import (
	"io"
	"net/http"
)

// FileInfo file information
type FileInfo struct {
	// file name
	Name string
	// file size
	Size int64
	// file the full path
	Path string
	// file extension
	Ext string
}

type (
	// FileNameHandle filename handling
	FileNameHandle func(base, filename string) string
	// FileSizeHandle file size limit handing
	FileSizeHandle func(size int64) bool
)

// Uploader file upload interface
type Uploader interface {
	// upload a single file
	Upload(r *http.Request, key string, fnh FileNameHandle, fsh FileSizeHandle) (*FileInfo, error)
	// upload multiple files
	UploadMulti(r *http.Request, key string, fnh FileNameHandle, fsh FileSizeHandle) ([]*FileInfo, error)
	// upload data stream file
	UploadReader(r io.Reader, fnh FileNameHandle, fsh FileSizeHandle) (*FileInfo, error)
}
