# fuh - upload handler library

[![License][License-Image]][License-Url] [![ReportCard][ReportCard-Image]][ReportCard-Url] [![Build][Build-Status-Image]][Build-Status-Url] [![Coverage][Coverage-Image]][Coverage-Url] [![GoDoc][GoDoc-Image]][GoDoc-Url]

## Quick Start

### Download and install

``` bash
go get github.com/LyricTian/fuh
```

### Create file `server.go`

``` go
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/LyricTian/fuh"
)

func main() {
	uploader := fuh.NewUploader(&fuh.Config{
		BasePath:  "attach",
		SizeLimit: 1 << 20,
	}, fuh.NewFileStore())

	http.HandleFunc("/fileupload", func(w http.ResponseWriter, r *http.Request) {

		ctx := fuh.NewFileNameContext(context.Background(), func(ci fuh.ContextInfo) string {
			return filepath.Join(ci.BasePath(), ci.FileName())
		})

		finfos, err := uploader.Upload(ctx, r, "file")
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

``` bash
$ go build server.go
$ ./server
```

## Features

* Custom file name
* Custom file size limit
* Support timeout handler
* Supports storage extensions

## MIT License

```
Copyright (c) 2017 LyricTian
```

[License-Url]: http://opensource.org/licenses/MIT
[License-Image]: https://img.shields.io/npm/l/express.svg
[Build-Status-Url]: https://travis-ci.org/LyricTian/fuh
[Build-Status-Image]: https://travis-ci.org/LyricTian/fuh.svg?branch=master
[ReportCard-Url]: https://goreportcard.com/report/github.com/LyricTian/fuh
[ReportCard-Image]: https://goreportcard.com/badge/github.com/LyricTian/fuh
[GoDoc-Url]: https://godoc.org/github.com/LyricTian/fuh
[GoDoc-Image]: https://godoc.org/github.com/LyricTian/fuh?status.svg
[Coverage-Url]: https://coveralls.io/github/LyricTian/fuh?branch=master
[Coverage-Image]: https://coveralls.io/repos/github/LyricTian/fuh/badge.svg?branch=master
