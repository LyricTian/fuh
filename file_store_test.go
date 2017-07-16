package fuh_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/LyricTian/fuh"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFileStore(t *testing.T) {
	basePath := "testdatas/"
	Convey("file storage test", t, func() {

		Convey("write data", func() {
			store := fuh.NewFileStore()
			buf := []byte("abc")
			filename := filepath.Join(basePath, "write.txt")

			err := store.Store(nil, filename, bytes.NewReader(buf), int64(len(buf)))
			So(err, ShouldBeNil)

			if err == nil {
				defer os.Remove(filename)
			}

			file, err := os.Open(filename)
			So(err, ShouldBeNil)
			defer file.Close()

			fbuf, err := ioutil.ReadAll(file)
			So(err, ShouldBeNil)
			So(string(fbuf), ShouldEqual, string(buf))
		})

		Convey("rewrite data", func() {
			filename := filepath.Join(basePath, "rewrite.txt")

			cfile, err := os.Create(filename)
			So(err, ShouldBeNil)

			defer os.Remove(filename)

			cfile.Write([]byte("123"))
			cfile.Close()

			buf := []byte("abc")
			store := &fuh.FileStore{Rewrite: true}
			err = store.Store(nil, filename, bytes.NewReader(buf), int64(len(buf)))
			So(err, ShouldBeNil)

			ofile, err := os.Open(filename)
			So(err, ShouldBeNil)
			defer ofile.Close()

			fbuf, err := ioutil.ReadAll(ofile)
			So(err, ShouldBeNil)
			So(string(fbuf), ShouldEqual, string(buf))
		})

	})
}
