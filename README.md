# Golang File Upload Handler

[![Build][Build-Status-Image]][Build-Status-Url] [![Codecov][codecov-image]][codecov-url] [![ReportCard][reportcard-image]][reportcard-url] [![GoDoc][godoc-image]][godoc-url] [![License][license-image]][license-url]

## Quick Start

### Download and install

```bash
go get -v github.com/LyricTian/fuh
```

### Create file `server.go`

```go
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/LyricTian/fuh"
)

func main() {
	upl := fuh.NewUploader(&fuh.Config{
		BasePath:  "attach",
		SizeLimit: 1 << 20,
	}, fuh.NewFileStore())

	http.HandleFunc("/fileupload", func(w http.ResponseWriter, r *http.Request) {

		ctx := fuh.NewFileNameContext(context.Background(), func(ci fuh.ContextInfo) string {
			return filepath.Join(ci.BasePath(), ci.FileName())
		})

		finfos, err := upl.Upload(ctx, r, "file")
		if err != nil {
			w.WriteHeader(500)
			return
		}
		json.NewEncoder(w).Encode(finfos)
	})

	http.ListenAndServe(":8080", nil)
}
```

### Build and run

```bash
$ go build server.go
$ ./server
```

## Features

- Custom file name
- Custom file size limit
- Supports storage extensions
- Context support

## MIT License

    Copyright (c) 2017 Lyric

[Build-Status-Url]: https://travis-ci.org/LyricTian/fuh
[Build-Status-Image]: https://travis-ci.org/LyricTian/fuh.svg?branch=master
[codecov-url]: https://codecov.io/gh/LyricTian/fuh
[codecov-image]: https://codecov.io/gh/LyricTian/fuh/branch/master/graph/badge.svg
[reportcard-url]: https://goreportcard.com/report/github.com/LyricTian/fuh
[reportcard-image]: https://goreportcard.com/badge/github.com/LyricTian/fuh
[godoc-url]: https://godoc.org/github.com/LyricTian/fuh
[godoc-image]: https://godoc.org/github.com/LyricTian/fuh?status.svg
[license-url]: http://opensource.org/licenses/MIT
[license-image]: https://img.shields.io/npm/l/express.svg