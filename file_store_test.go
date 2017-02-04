package fuh_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/LyricTian/fuh"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileStore(t *testing.T) {
	Convey("Local file store", t, func() {
		filename := "testdatas/upload/store.txt"
		os.Remove(filename)

		store := fuh.NewFileStore()
		ds := "foo"
		buf := bytes.NewBuffer([]byte(ds))

		err := store.Store(filename, buf, int64(buf.Len()))
		So(err, ShouldBeNil)

		file, err := os.Open(filename)
		So(err, ShouldBeNil)
		So(file, ShouldNotBeNil)

		fbuf := new(bytes.Buffer)
		io.Copy(fbuf, file)
		file.Close()
		So(fbuf.String(), ShouldEqual, ds)

		Convey("Existing file failed", func() {
			store := fuh.NewFileStore()
			buf := bytes.NewBuffer([]byte("123"))
			err := store.Store(filename, bytes.NewBuffer([]byte("123")), int64(buf.Len()))
			So(err, ShouldNotBeNil)
		})

		Convey("Rewriting the file", func() {
			store := fuh.NewFileStore(&fuh.FileStoreConfig{Rewrite: true})
			ds := "bar"
			buf := bytes.NewBuffer([]byte(ds))

			err := store.Store(filename, buf, int64(buf.Len()))
			So(err, ShouldBeNil)

			file, err := os.Open(filename)
			So(err, ShouldBeNil)
			So(file, ShouldNotBeNil)

			fbuf := new(bytes.Buffer)
			io.Copy(fbuf, file)
			file.Close()
			So(fbuf.String(), ShouldEqual, ds)
		})
	})
}
