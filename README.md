# File upload handler

[![License][License-Image]][License-Url] [![ReportCard][ReportCard-Image]][ReportCard-Url] [![Build][Build-Status-Image]][Build-Status-Url] [![Coverage][Coverage-Image]][Coverage-Url] [![GoDoc][GoDoc-Image]][GoDoc-Url]

## Quick Start

### Download and install

``` bash
$ go get github.com/LyricTian/fuh
```

### Create file `server.go`

``` go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/LyricTian/fuh"
)

func main() {

	fstore := fuh.NewFileStore()
	fupl := fuh.NewUploader(fstore, &fuh.UploadConfig{BasePath: "files", SizeLimit: 1024 * 1024})

	http.HandleFunc("/fileupload", func(w http.ResponseWriter, r *http.Request) {
		finfo, err := fupl.Upload(r, "file", nil, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(finfo)
	})

	http.ListenAndServe(":8080", nil)
}

```

### Build and run

``` bash
$ go build server.go
$ ./server
```

## Test

``` bash
$ go test -v
```

## MIT License

```
Copyright (c) 2016 LyricTian
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
