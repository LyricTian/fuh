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
	maxMemory  int64       = 32 << 20 // 32M
	fnReplacer             = strings.NewReplacer("\\", "_")
	storeKey   interface{} = "Store"
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

func (uh *uploadHandle) setConfig(cfg *UploadConfig) {
	if cfg == nil {
		return
	}
	uh.cfg = cfg
}

func (uh *uploadHandle) getConfig() *UploadConfig {
	if uh.cfg == nil {
		uh.cfg = &UploadConfig{}
	}

	if uh.cfg.MaxMemory == 0 {
		uh.cfg.MaxMemory = maxMemory
	}

	return uh.cfg
}

func (uh *uploadHandle) setStore(store Storer) {
	if store == nil {
		return
	}
	uh.store = store
}

func (uh *uploadHandle) getStore() Storer {
	if uh.store == nil {
		uh.store = NewFileStore()
	}
	return uh.store
}

func (uh *uploadHandle) randName() (name string) {
	buf := make([]byte, 16)
	n, _ := rand.Read(buf)
	name = hex.EncodeToString(buf[:n])
	return
}

func (uh *uploadHandle) parseHeaders(r *http.Request, key string) (headers []*multipart.FileHeader, err error) {
	if r.MultipartForm == nil {
		if err = r.ParseMultipartForm(uh.getConfig().MaxMemory); err != nil {
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
		basePath = uh.getConfig().BasePath
	)

	if fnh != nil {
		fullName = fnh(basePath, fileName)
	}

	if fullName == "" {
		fullName = filepath.Join(basePath, fileName)
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
	sizeLimit := uh.getConfig().SizeLimit

	if (fsh != nil && !fsh(fileSize)) ||
		(sizeLimit != 0 && fileSize > sizeLimit) {
		err = ErrFileTooLarge
		return
	}

	err = uh.getStore().Store(fullName, file, fileSize)
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
	var (
		fullName string
		basePath = uh.getConfig().BasePath
	)

	if fnh != nil {
		fullName = fnh(basePath, "")
	}

	if fullName == "" {
		fullName = filepath.Join(basePath, uh.randName())
	}

	buf := new(bytes.Buffer)
	fileSize, _ := buf.ReadFrom(r)

	if fileSize == 0 {
		err = ErrMissingFile
		return
	}

	sizeLimit := uh.getConfig().SizeLimit
	if (fsh != nil && !fsh(fileSize)) ||
		(sizeLimit != 0 && fileSize > sizeLimit) {
		err = ErrFileTooLarge
		return
	}

	err = uh.getStore().Store(fullName, buf, fileSize)
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
