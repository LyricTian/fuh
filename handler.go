package fuh

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

var (
	// ErrMissingFile no such file
	ErrMissingFile = errors.New("no such file")
	// ErrFileTooLarge file too large
	ErrFileTooLarge = errors.New("file too large")
)

// Config basic configuration
type Config struct {
	BasePath  string
	SizeLimit int64
	MaxMemory int64
}

// NewUploader create a file upload interface
func NewUploader(cfg *Config, store Storer) Uploader {
	return &uploadHandle{
		cfg:   cfg,
		store: store,
	}
}

func newUploader() *uploadHandle {
	return &uploadHandle{}
}

type uploadHandle struct {
	cfg   *Config
	store Storer
}

func (u *uploadHandle) Upload(ctx context.Context, r *http.Request, key string) ([]FileInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if u.store == nil {
		u.store = NewFileStore()
	}

	if r.MultipartForm == nil {
		err := r.ParseMultipartForm(u.maxMemory())
		if err != nil {
			return nil, err
		}
	}

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, ErrMissingFile
	}

	var finfos []FileInfo
	for _, file := range r.MultipartForm.File[key] {
		err := u.uploadDo(ctx, r, file, func(finfo FileInfo) {
			finfos = append(finfos, finfo)
		})
		if err != nil {
			return finfos, err
		}
	}
	return finfos, nil
}

func (u *uploadHandle) config() *Config {
	if c := u.cfg; c != nil {
		return c
	}
	return &Config{}
}

func (u *uploadHandle) maxMemory() int64 {
	if mm := u.config().MaxMemory; mm > 0 {
		return mm
	}
	return defaultMaxMemory
}

func (u *uploadHandle) uploadDo(ctx context.Context, r *http.Request, fheader *multipart.FileHeader, f func(FileInfo)) error {
	c := make(chan error, 1)

	go func() {
		file, err := fheader.Open()
		if err != nil {
			c <- err
			return
		}
		defer file.Close()

		fsize, err := u.fileSize(file)
		if err != nil {
			c <- err
			return
		}

		var ctxInfo ContextInfo = &contextInfo{
			basePath:   u.config().BasePath,
			fileName:   fheader.Filename,
			fileSize:   fsize,
			fileHeader: fheader.Header,
			req:        r,
		}

		// upload file size limit
		if h, ok := FromFileSizeLimitContext(ctx); ok {
			if !h(ctxInfo) {
				c <- ErrFileTooLarge
				return
			}
		} else if sl := u.config().SizeLimit; sl > 0 && fsize > sl {
			c <- ErrFileTooLarge
			return
		}

		var fullName string
		if h, ok := FromFileNameContext(ctx); ok {
			fullName = h(ctxInfo)
		} else {
			fullName = filepath.Join(ctxInfo.BasePath(), ctxInfo.FileName())
		}

		err = u.store.Store(ctx, fullName, file, fsize)
		if err != nil {
			c <- err
			return
		}

		var fInfo FileInfo = &fileInfo{
			fullName: fullName,
			size:     fsize,
		}

		f(fInfo)
		c <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}

// get the size of the uploaded file
func (u *uploadHandle) fileSize(file multipart.File) (int64, error) {
	var size int64

	if fsize, ok := file.(fsize); ok {
		size = fsize.Size()
	} else if fstat, ok := file.(fstat); ok {
		stat, err := fstat.Stat()
		if err != nil {
			return 0, err
		}
		size = stat.Size()
	}

	return size, nil
}

type fsize interface {
	Size() int64
}

type fstat interface {
	Stat() (os.FileInfo, error)
}

type fileInfo struct {
	fullName string
	size     int64
}

func (fi *fileInfo) FullName() string {
	return fi.fullName
}

func (fi *fileInfo) Size() int64 {
	return fi.size
}
