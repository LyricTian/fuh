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
func NewFileStore() Storer {
	return &FileStore{}
}

// FileStore file storage
type FileStore struct {
	// rewrite the existing file
	Rewrite bool
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

	if ctx == nil {
		ctx = context.Background()
	}

	c := make(chan error, 1)
	ctxDone := make(chan struct{})

	go func() {
		exists, err := f.exists(filename)
		if err != nil {
			c <- err
			return
		} else if exists {
			if !f.Rewrite {
				c <- ErrFileExists
				return
			}
			os.Remove(filename)
		}

		dir := filepath.Dir(filename)
		if dir != "" {
			exists, err = f.exists(dir)
			if err != nil {
				c <- err
				return
			} else if !exists {
				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					c <- err
					return
				}
			}
		}

		file, err := os.Create(filename)
		if err != nil {
			c <- err
			return
		}
		defer file.Close()

		_, err = io.CopyN(file, data, size)
		if err != nil {
			c <- err
			return
		}

		select {
		case <-ctxDone:
			os.Remove(filename)
		default:
		}

		c <- nil
	}()

	select {
	case <-ctx.Done():
		close(ctxDone)
		return ctx.Err()
	case err := <-c:
		return err
	}
}
