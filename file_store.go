package fuh

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	// ErrNoData no data is stored
	ErrNoData = errors.New("no data is stored")
	// ErrFileExists file already exists
	ErrFileExists = errors.New("file already exists")
)

// NewFileStore create a file store
func NewFileStore() *FileStore {
	return &FileStore{}
}

// NewFileStoreWithBasePath create a file store with base path
func NewFileStoreWithBasePath(basePath string) *FileStore {
	return &FileStore{
		BasePath: basePath,
	}
}

// FileStore file storage
type FileStore struct {
	// rewrite the existing file
	Rewrite  bool
	BasePath string
}

func (f *FileStore) exists(filename string) (exists bool, err error) {
	exists = true
	_, verr := os.Stat(filename)
	if verr != nil {
		if os.IsNotExist(verr) {
			exists = false
			return
		}
		err = verr
	}
	return
}

// Store store data to a local file
func (f *FileStore) Store(ctx context.Context, filename string, data io.Reader, size int64) error {
	if filename == "" || data == nil || size == 0 {
		return ErrNoData
	}

	if f.BasePath != "" {
		filename = filepath.Join(f.BasePath, filename)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	exists, err := f.exists(filename)
	if err != nil {
		return err
	} else if exists {
		if !f.Rewrite {
			return ErrFileExists
		}
		os.Remove(filename)
	}

	dir := filepath.Dir(filename)
	if dir != "" {
		if exists, err := f.exists(dir); err != nil {
			return err
		} else if !exists {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.CopyN(file, data, size)
	return err
}
