package main

import (
	"io"
	"os"
	"path"
	"time"
)

// TODO: This becomes a FIFO since it's available on every target but plan9
type Reader struct {
	io.ReadCloser
}

func NewReader(name string) (*Reader, error) {
	os.MkdirAll(path.Dir(name), 0755)
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0755)
	if err != nil {
		return &Reader{}, err
	}
	if _, err := f.Seek(0, 2); err != nil {
		return &Reader{f}, err
	}
	return &Reader{f}, err
}

func (r *Reader) Read(p []byte) (n int, err error) {
	for {
		n, err := r.ReadCloser.Read(p)
		if n > 0 {
			if n > 0 {
				return n, nil
			} else if err != io.EOF {
				return n, err
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}
