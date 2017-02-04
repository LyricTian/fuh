package fuh

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	maxMemory  int64 = 32 << 20 // 32M
	fnReplacer       = strings.NewReplacer("\\", "_")
)

// UploadConfig upload the configuration parameters
type UploadConfig struct {
	// the base path of storing files
	BasePath string
	// file size limit(the default is not limit)
	SizeLimit int64
	// The whole request body is parsed and up to a total of maxMemory bytes of
	// its file parts are stored in memory, with the remainder stored on
	// disk in temporary files(the default of 32M).
	MaxMemory int64
}

// NewUploader create a file uploader
func NewUploader(store Storer, cfgs ...*UploadConfig) Uploader {
	if store == nil {
		panic("invalid store")
	}

	cfg := &UploadConfig{}
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}

	if cfg.MaxMemory == 0 {
		cfg.MaxMemory = maxMemory
	}

	return &uploadHandle{
		store: store,
		cfg:   cfg,
	}
}

type uploadHandle struct {
	store Storer
	cfg   *UploadConfig
}

func (uh *uploadHandle) randName() (name string) {
	buf := make([]byte, 16)
	n, _ := rand.Read(buf)
	name = hex.EncodeToString(buf[:n])
	return
}

func (uh *uploadHandle) parseHeaders(r *http.Request, key string) (headers []*multipart.FileHeader, err error) {
	if r.MultipartForm == nil {
		if err = r.ParseMultipartForm(uh.cfg.MaxMemory); err != nil {
			return
		}
	}

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if v := r.MultipartForm.File[key]; len(v) > 0 {
			headers = v
			return
		}
	}

	err = ErrMissingFile
	return
}

type fsize interface {
	Size() int64
}

func (uh *uploadHandle) fileHandle(fileheader *multipart.FileHeader, fnh FileNameHandle, fsh FileSizeHandle) (info *FileInfo, err error) {
	var (
		fullName string
		fileName = fnReplacer.Replace(fileheader.Filename)
	)

	if fnh != nil {
		fullName = fnh(uh.cfg.BasePath, fileName)
	}

	if fullName == "" {
		fullName = filepath.Join(uh.cfg.BasePath, fileName)
	}

	file, err := fileheader.Open()
	if err != nil {
		return
	}
	defer file.Close()

	size, ok := file.(fsize)
	if !ok || size.Size() == 0 {
		err = ErrMissingFile
		return
	}

	fileSize := size.Size()
	if (fsh != nil && !fsh(fileSize)) ||
		(uh.cfg.SizeLimit != 0 && fileSize > uh.cfg.SizeLimit) {
		err = ErrFileTooLarge
		return
	}

	err = uh.store.Store(fullName, file, fileSize)
	if err != nil {
		return
	}

	info = &FileInfo{
		Name: filepath.Base(fullName),
		Ext:  filepath.Ext(fileName),
		Size: fileSize,
		Path: fullName,
	}

	return
}

func (uh *uploadHandle) Upload(r *http.Request, key string, fnh FileNameHandle, fsh FileSizeHandle) (info *FileInfo, err error) {
	headers, err := uh.parseHeaders(r, key)
	if err != nil {
		return
	}
	info, err = uh.fileHandle(headers[0], fnh, fsh)
	return
}

func (uh *uploadHandle) UploadMulti(r *http.Request, key string, fnh FileNameHandle, fsh FileSizeHandle) (infos []*FileInfo, err error) {
	headers, err := uh.parseHeaders(r, key)
	if err != nil {
		return
	}

	for _, header := range headers {
		info, verr := uh.fileHandle(header, fnh, fsh)
		if verr != nil {
			err = verr
			return
		}
		infos = append(infos, info)
	}

	return
}

func (uh *uploadHandle) UploadReader(r io.Reader, fnh FileNameHandle, fsh FileSizeHandle) (info *FileInfo, err error) {
	var fullName string
	if fnh != nil {
		fullName = fnh(uh.cfg.BasePath, "")
	}

	if fullName == "" {
		fullName = filepath.Join(uh.cfg.BasePath, uh.randName())
	}

	buf := new(bytes.Buffer)
	fileSize, _ := buf.ReadFrom(r)

	if fileSize == 0 {
		err = ErrMissingFile
		return
	}

	if (fsh != nil && !fsh(fileSize)) ||
		(uh.cfg.SizeLimit != 0 && fileSize > uh.cfg.SizeLimit) {
		err = ErrFileTooLarge
		return
	}

	err = uh.store.Store(fullName, buf, fileSize)
	if err != nil {
		return
	}

	info = &FileInfo{
		Name: filepath.Base(fullName),
		Ext:  filepath.Ext(fullName),
		Size: fileSize,
		Path: fullName,
	}

	return
}
