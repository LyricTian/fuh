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

	var infos []FileInfo
	for _, file := range r.MultipartForm.File[key] {
		info, err := u.uploadDo(ctx, r, file)
		if err != nil {
			return infos, err
		}
		infos = append(infos, info)
	}
	return infos, nil
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

func (u *uploadHandle) uploadDo(ctx context.Context, r *http.Request, fheader *multipart.FileHeader) (FileInfo, error) {
	file, err := fheader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	size, err := u.fileSize(file)
	if err != nil {
		return nil, err
	}

	ctxInfo := &contextInfo{
		basePath:   u.config().BasePath,
		fileName:   fheader.Filename,
		fileSize:   size,
		fileHeader: fheader.Header,
		req:        r,
	}
	ctx = NewContextInfoContext(ctx, ctxInfo)

	if h, ok := FromFileSizeLimitContext(ctx); ok {
		if !h(ctxInfo) {
			return nil, ErrFileTooLarge
		}
	} else if sl := u.config().SizeLimit; sl > 0 && size > sl {
		return nil, ErrFileTooLarge
	}

	var fullName string
	if h, ok := FromFileNameContext(ctx); ok {
		fullName = h(ctxInfo)
	} else {
		fullName = filepath.Join(ctxInfo.BasePath(), ctxInfo.FileName())
	}

	err = u.store.Store(ctx, fullName, file, size)
	if err != nil {
		return nil, err
	}

	return &fileInfo{
		fullName: fullName,
		name:     ctxInfo.FileName(),
		size:     size,
	}, nil
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
	name     string
	size     int64
}

func (fi *fileInfo) FullName() string {
	return fi.fullName
}

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) Size() int64 {
	return fi.size
}
