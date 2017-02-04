package fuh

import (
	"io"
)

// Storer upload data storage interface
type Storer interface {
	// the file data storage
	Store(filename string, r io.Reader, size int64) error
}
