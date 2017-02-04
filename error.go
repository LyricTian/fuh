package fuh

import "errors"

var (
	// ErrMissingFile no such file
	ErrMissingFile = errors.New("no such file")
	// ErrFileTooLarge file too large
	ErrFileTooLarge = errors.New("file too large")
	// ErrFileExists file already exists
	ErrFileExists = errors.New("file already exists")
)
