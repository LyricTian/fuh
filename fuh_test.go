package fuh_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/LyricTian/fuh"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFileUpload(t *testing.T) {
	basePath := "testdatas/"
	filename := "single_test.txt"
	buf := []byte("abc")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Convey("single file upload test", t, func() {
			uploader := fuh.NewUploader(&fuh.Config{BasePath: basePath}, fuh.NewFileStore())

			fileInfos, err := uploader.Upload(nil, r, "file")
			So(err, ShouldBeNil)
			So(len(fileInfos), ShouldEqual, 1)
			So(fileInfos[0].Size(), ShouldEqual, len(buf))
			So(fileInfos[0].FullName(), ShouldEqual, filepath.Join(basePath, filename))

			defer os.Remove(fileInfos[0].FullName())

			file, err := os.Open(fileInfos[0].FullName())
			So(err, ShouldBeNil)
			defer file.Close()

			buf, err := ioutil.ReadAll(file)
			So(err, ShouldBeNil)
			So(string(buf), ShouldEqual, string(buf))
		})
	}))
	defer srv.Close()

	err := postFile(srv.URL, buf, filename)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestCustomFileName(t *testing.T) {
	basePath := "testdatas/"
	filename := "filename_test.txt"
	buf := []byte("abc")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Convey("custom file name test", t, func() {
			uploader := fuh.NewUploader(&fuh.Config{BasePath: basePath}, fuh.NewFileStore())

			nfilename := "nfilename_test.txt"
			ctx := fuh.NewFileNameContext(context.Background(), func(ci fuh.ContextInfo) string {
				return filepath.Join(ci.BasePath(), nfilename)
			})

			fileInfos, err := uploader.Upload(ctx, r, "file")
			So(err, ShouldBeNil)
			So(len(fileInfos), ShouldEqual, 1)
			So(fileInfos[0].Size(), ShouldEqual, len(buf))
			So(fileInfos[0].FullName(), ShouldEqual, filepath.Join(basePath, nfilename))

			defer os.Remove(fileInfos[0].FullName())

			file, err := os.Open(fileInfos[0].FullName())
			So(err, ShouldBeNil)
			defer file.Close()

			buf, err := ioutil.ReadAll(file)
			So(err, ShouldBeNil)
			So(string(buf), ShouldEqual, string(buf))
		})
	}))
	defer srv.Close()

	err := postFile(srv.URL, buf, filename)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestFileSizeLimit(t *testing.T) {
	basePath := "testdatas/"
	filename := "filesizelimit_test.txt"
	buf := []byte("abc")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Convey("file size limit test", t, func() {
			uploader := fuh.NewUploader(&fuh.Config{BasePath: basePath}, fuh.NewFileStore())

			ctx := fuh.NewFileSizeLimitContext(context.Background(), func(ci fuh.ContextInfo) bool {
				return ci.FileSize() < 3
			})

			fileInfos, err := uploader.Upload(ctx, r, "file")
			So(fileInfos, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, fuh.ErrFileTooLarge)
		})
	}))
	defer srv.Close()

	err := postFile(srv.URL, buf, filename)
	if err != nil {
		t.Error(err.Error())
	}
}

func postFile(targetURL string, data []byte, filenames ...string) (err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	for _, filename := range filenames {
		fileWriter, verr := bodyWriter.CreateFormFile("file", filename)
		if verr != nil {
			err = verr
			return
		}
		fileWriter.Write(data)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	http.Post(targetURL, contentType, bodyBuf)

	return
}
