package main

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/snappy"
	"github.com/klauspost/pgzip"
	"github.com/pierrec/lz4"
	"github.com/ulikunitz/xz"
)

func mustGzip(rd io.Reader) io.Reader {
	r, err := gzip.NewReader(rd)
	if err != nil {
		panic(err)
	}
	return r
}
func mustPGzip(rd io.Reader) io.Reader {
	r, err := pgzip.NewReader(rd)
	if err != nil {
		panic(err)
	}
	return r
}

func BenchmarkGunzip(b *testing.B) {
	decompress(b, "gzip", mustGzip)
}
func BenchmarkPGunzip(b *testing.B) {
	decompress(b, "gzip", mustPGzip)
}
func BenchmarkBunzip(b *testing.B) {
	decompress(b, "bzip2", bzip2.NewReader)
}
func BenchmarkSnappy(b *testing.B) {
	decompress(b, "snappy", func(f io.Reader) io.Reader { return snappy.NewReader(f) })
}
func BenchmarkLZ4(b *testing.B) {
	decompress(b, "lz4", func(f io.Reader) io.Reader { return lz4.NewReader(f) })
}
func BenchmarkXZ(b *testing.B) {
	decompress(b, "xz", func(f io.Reader) io.Reader {
		r, err := xz.NewReader(f)
		if err != nil {
			panic(err)
		}
		return r
	})
}
func BenchmarkRaw(b *testing.B) {
	decompress(b, "", func(f io.Reader) io.Reader { return f })
}

/*
func init() {
	err := WriteLZ4()
	if err != nil {
		panic(err)
	}
}
*/

func WriteSnappy() error {
	return writeGeneric("snappy", func(w io.Writer) io.WriteCloser { return snappy.NewWriter(w) })
}
func WriteLZ4() error {
	return writeGeneric("lz4", func(w io.Writer) io.WriteCloser { return lz4.NewWriter(w) })
}

func writeGeneric(suffix string, wrap func(io.Writer) io.WriteCloser) error {
	src, err := os.Open("testdata/100m")
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create("testdata/100m." + suffix)
	if err != nil {
		return err
	}
	defer dst.Close()
	w := wrap(dst)
	_, err = io.Copy(w, src)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

const flen = 100 * 1024 * 1024

func decompress(b *testing.B, suffix string, wrap func(io.Reader) io.Reader) {
	b.SetBytes(flen)
	var filename = "testdata/100m"
	if suffix != "" {
		filename += "." + suffix
	}
	for i := 0; i < b.N; i++ {
		f, err := os.Open(filename)
		if err != nil {
			b.Fatal("open", err)
		}
		defer f.Close()
		rd := wrap(f)
		n, err := io.Copy(ioutil.Discard, rd)
		if err != nil {
			b.Fatal("copy", err)
		}
		if n != flen {
			b.Fatal("short", n, "expected", flen)
		}
	}
}
