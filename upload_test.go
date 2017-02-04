package fuh_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/LyricTian/fuh"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	fbase = "testdatas"
	fkey  = "file"
	upl   = fuh.NewUploader(fuh.NewFileStore(), &fuh.UploadConfig{
		BasePath: filepath.Join(fbase, "upload"),
	})
)

func fnHandle(base, filename string) string {
	return filepath.Join(base, randName(), filename)
}

func TestUpload(t *testing.T) {
	filename := "file_1.txt"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		Convey("Single file upload", t, func() {
			finfo, err := upl.Upload(r, fkey, fnHandle, nil)
			So(err, ShouldBeNil)
			So(finfo, ShouldNotBeNil)
			So(finfo.Name, ShouldEqual, filename)
		})

	}))
	defer srv.Close()

	postFile(srv.URL, filename)
}

func TestMultiUpload(t *testing.T) {
	filenames := []string{"file_1.txt", "file_2.txt"}
	rname := randName()

	var fnHandle = func(base, filename string) string {
		return filepath.Join(base, rname, filename)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		Convey("Multiple files upload", t, func() {
			finfos, err := upl.UploadMulti(r, fkey, fnHandle, nil)
			So(err, ShouldBeNil)
			So(finfos, ShouldNotBeNil)
			So(len(finfos), ShouldEqual, len(filenames))
		})

	}))
	defer srv.Close()

	postFile(srv.URL, filenames...)
}

func TestUploadSizeLimit(t *testing.T) {
	filename := "file_2.txt"

	var fsHandle = func(size int64) bool {
		return size <= 5
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		Convey("File upload size limit", t, func() {
			finfo, err := upl.Upload(r, fkey, fnHandle, fsHandle)
			So(err, ShouldNotBeNil)
			So(finfo, ShouldBeNil)
		})

	}))
	defer srv.Close()

	postFile(srv.URL, filename)
}

func TestUploadReader(t *testing.T) {
	Convey("File reader upload", t, func() {
		filename := "file_2.txt"

		file, err := os.Open(filepath.Join(fbase, filename))
		So(err, ShouldBeNil)
		So(file, ShouldNotBeNil)
		defer file.Close()

		var fnHandle = func(base, fname string) string {
			return filepath.Join(base, randName(), filename)
		}

		finfo, err := upl.UploadReader(file, fnHandle, nil)
		So(err, ShouldBeNil)
		So(finfo, ShouldNotBeNil)
		So(finfo.Name, ShouldEqual, filename)
	})
}

func randName() (name string) {
	buf := make([]byte, 16)
	n, _ := rand.Read(buf)
	name = hex.EncodeToString(buf[:n])
	return
}

func postFile(targetURL string, filenames ...string) (err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	for _, filename := range filenames {
		fileWriter, verr := bodyWriter.CreateFormFile(fkey, filepath.Base(filename))
		if verr != nil {
			err = verr
			return
		}

		fh, verr := os.Open(filepath.Join(fbase, filename))
		if verr != nil {
			err = verr
			return
		}
		io.Copy(fileWriter, fh)
		fh.Close()
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	http.Post(targetURL, contentType, bodyBuf)

	return
}
