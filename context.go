package fuh

import (
	"context"
	"net/http"
	"net/textproto"
)

type (
	fileSizeLimitKey struct{}
	fileNameKey      struct{}
	contextInfoKey   struct{}

	// ContextInfo the context information
	ContextInfo interface {
		BasePath() string
		FileName() string
		FileSize() int64
		FileHeader() textproto.MIMEHeader
		Request() *http.Request
	}

	// FileSizeLimitHandle file size limit
	FileSizeLimitHandle func(ci ContextInfo) bool

	// FileNameHandle the file name
	FileNameHandle func(ci ContextInfo) string
)

// NewFileSizeLimitContext returns a new Context that carries value fsl.
func NewFileSizeLimitContext(ctx context.Context, fsl FileSizeLimitHandle) context.Context {
	return context.WithValue(ctx, fileSizeLimitKey{}, fsl)
}

// FromFileSizeLimitContext returns the FileSizeLimitHandle value stored in ctx, if any.
func FromFileSizeLimitContext(ctx context.Context) (FileSizeLimitHandle, bool) {
	handle, ok := ctx.Value(fileSizeLimitKey{}).(FileSizeLimitHandle)
	return handle, ok
}

// NewFileNameContext returns a new Context that carries value fn.
func NewFileNameContext(ctx context.Context, fn FileNameHandle) context.Context {
	return context.WithValue(ctx, fileNameKey{}, fn)
}

// FromFileNameContext returns the FileNameHandle value stored in ctx, if any.
func FromFileNameContext(ctx context.Context) (FileNameHandle, bool) {
	handle, ok := ctx.Value(fileNameKey{}).(FileNameHandle)
	return handle, ok
}

// NewContextInfoContext returns a new Context that context information ci.
func NewContextInfoContext(ctx context.Context, ci ContextInfo) context.Context {
	return context.WithValue(ctx, contextInfoKey{}, ci)
}

// FromContextInfoContext returns the ContextInfo value stored in ctx, if any.
func FromContextInfoContext(ctx context.Context) (ContextInfo, bool) {
	info, ok := ctx.Value(contextInfoKey{}).(ContextInfo)
	return info, ok
}

type contextInfo struct {
	basePath   string
	fileName   string
	fileSize   int64
	fileHeader textproto.MIMEHeader
	req        *http.Request
}

func (ci *contextInfo) BasePath() string {
	return ci.basePath
}

func (ci *contextInfo) FileName() string {
	return ci.fileName
}

func (ci *contextInfo) FileSize() int64 {
	return ci.fileSize
}

func (ci *contextInfo) FileHeader() textproto.MIMEHeader {
	return ci.fileHeader
}

func (ci *contextInfo) Request() *http.Request {
	return ci.req
}
