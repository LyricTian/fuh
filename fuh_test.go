package fuh_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LyricTian/fuh"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFileUpload(t *testing.T) {
	basePath := "testdatas/"
	filename := "single_test.txt"
	buf := []byte("abc")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Convey("single file upload test", t, func() {
			config := &fuh.Config{BasePath: basePath, SizeLimit: 1 << 20, MaxMemory: 10 << 20}

			fuh.SetConfig(config)
			fuh.SetStore(fuh.NewFileStore())

			fileInfos, err := fuh.Upload(nil, r, "file")
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

	postFile(srv.URL, buf, filename)
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

	postFile(srv.URL, buf, filename)
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

	postFile(srv.URL, buf, filename)
}

func TestUploadTimeout(t *testing.T) {
	basePath := "testdatas/"
	filename := "uploadtimeout_test.txt"
	buf := []byte("abc")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Convey("file upload timeout test", t, func() {
			uploader := fuh.NewUploader(&fuh.Config{BasePath: basePath}, fuh.NewFileStore())

			ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*10)
			defer cancel()

			fileInfos, err := uploader.Upload(ctx, r, "file")
			So(fileInfos, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, context.DeadlineExceeded.Error())
		})
	}))
	defer srv.Close()

	postFile(srv.URL, buf, filename)
}

func TestFileSize(t *testing.T) {
	basePath := "testdatas/"
	filename := "filesize_test.txt"
	buf := make([]byte, 1024)
	for i := 0; i < 1024; i++ {
		buf[i] = '0'
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Convey("file size test", t, func() {
			uploader := fuh.NewUploader(&fuh.Config{BasePath: basePath, MaxMemory: 128}, fuh.NewFileStore())

			defer os.Remove(filepath.Join(basePath, filename))

			fileInfos, err := uploader.Upload(context.Background(), r, "file")
			So(err, ShouldBeNil)
			So(len(fileInfos), ShouldEqual, 1)
			So(fileInfos[0].Size(), ShouldEqual, len(buf))

		})
	}))
	defer srv.Close()

	postFile(srv.URL, buf, filename)
}

func postFile(targetURL string, data []byte, filenames ...string) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	for _, filename := range filenames {
		fileWriter, err := bodyWriter.CreateFormFile("file", filename)
		if err != nil {
			log.Println(err.Error())
			return
		}
		fileWriter.Write(data)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	http.Post(targetURL, contentType, bodyBuf)

	return
}
