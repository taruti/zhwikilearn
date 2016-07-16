package main

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func mustGzip(rd io.Reader) io.Reader {
	r, err := gzip.NewReader(rd)
	if err != nil {
		panic(err)
	}
	return r
}

func BenchmarkGunzip(b *testing.B) {
	decompress(b, "gzip", mustGzip)
}
func BenchmarkBunzip(b *testing.B) {
	decompress(b, "bzip2", bzip2.NewReader)
}

const flen = 100 * 1024 * 1024

func decompress(b *testing.B, suffix string, wrap func(io.Reader) io.Reader) {
	b.SetBytes(flen)
	for i := 0; i < b.N; i++ {
		f, err := os.Open("testdata/100m." + suffix)
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
