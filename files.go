package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

type fakefile struct {
	event  chan string
	name   string
	offset int64
}

func (f *fakefile) ReadAt(p []byte, off int64) (int, error) {
	n := copy(p[off:], "test\n")
	return n, nil
}

// Called on input and ctl
func (f *fakefile) Write(p []byte, off int64) (int, error) {
	f.event <- string(p[off:])
	return 0, nil
}

func (f *fakefile) Close() error {
	return nil
}

func (f *fakefile) size() int64 {
	if f.name == "/" {
		return 0
	}
	return int64(len(fmt.Sprint(f)))
}

type stat struct {
	name string
	file *fakefile
}

func (s *stat) Name() string     { return s.name }
func (s *stat) Sys() interface{} { return s.file }

func (s *stat) ModTime() time.Time {
	return time.Now().Truncate(time.Hour)
}

// We have only one directory, so return that
func (s *stat) IsDir() bool {
	return (s.name == "/")
}

// Again, only root directory so we can safely optimize
func (s *stat) Mode() os.FileMode {
	switch s.name {
	case "/":
		return os.ModeDir | 0755
	case "input", "ctl":
		return 0666
	}
	return 0444
}

func (s *stat) Size() int64 {
	return s.file.size()
}

type dir struct {
	c    chan stat
	done chan struct{}
}

func mkdir(st *State) *dir {
	c := make(chan stat, 10)
	done := make(chan struct{})
	go func() {
		for name, show := range st.show {
			if show {
				select {
				case c <- stat{name: name, file: &fakefile{name: name}}:
				case <-done:
					break
				}
			}
		}
		close(c)
	}()
	return &dir{
		c:    c,
		done: done,
	}
}

// This is fine for our needs
func (d *dir) Readdir(n int) ([]os.FileInfo, error) {
	var err error
	fi := make([]os.FileInfo, 0, 10)
	for i := 0; i < n; i++ {
		s, ok := <-d.c
		if !ok {
			err = io.EOF
			break
		}
		fi = append(fi, &s)
	}
	return fi, err
}

func (d *dir) Close() error {
	close(d.done)
	return nil
}
