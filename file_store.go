package fuh

import (
	"io"
	"os"
	"path/filepath"
)

// FileStoreConfig file store configuration parameters
type FileStoreConfig struct {
	// rewrite the existing files
	Rewrite bool
}

// NewFileStore create a file store
func NewFileStore(cfgs ...*FileStoreConfig) Storer {
	cfg := &FileStoreConfig{}
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}

	return &fileStore{
		cfg: cfg,
	}
}

type fileStore struct {
	cfg *FileStoreConfig
}

func (fs *fileStore) exists(name string) (exists bool, err error) {
	exists = true
	_, verr := os.Stat(name)
	if verr != nil {
		if os.IsNotExist(verr) {
			exists = false
			return
		}
		err = verr
	}
	return
}

func (fs *fileStore) Store(filename string, r io.Reader, size int64) (err error) {
	exists, err := fs.exists(filename)
	if err != nil {
		return
	} else if exists {
		if !fs.cfg.Rewrite {
			err = ErrFileExists
			return
		}
		os.Remove(filename)
	}

	dir := filepath.Dir(filename)
	if dir != "" {
		exists, err = fs.exists(dir)
		if err != nil {
			return
		} else if !exists {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return
			}
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()

	io.CopyN(file, r, size)
	return
}
