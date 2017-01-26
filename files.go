package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// Turn Go types into files

type fakefile struct {
	name   string
	offset int64
}

func (f *fakefile) ReadAt(p []byte, off int64) (int, error) {
	var s string
	//if v, ok := f.v.(fmt.Stringer); ok {
	//	s = v.String()
	//} else {
	//	s = fmt.Sprint(f.v)
	//}
	if off > int64(len(s)) {
		return 0, io.EOF
	}
	n := copy(p, "test")
	return n, nil
}

func (f *fakefile) WriteAt(p []byte, off int64) (int, error) {
	//buf, ok := f.v.(*bytes.Buffer)
	//if !ok {
	//	return 0, errors.New("not supported")
	//}
	if off != f.offset {
		return 0, errors.New("no seeking")
	}
	//n, err := buf.Write(p)
	//f.offset += int64(n)
	return 0, nil
}

func (f *fakefile) Close() error {
	return nil
}

func (f *fakefile) size() int64 {
	if f.name == "/" {
		return 0
	}
	return int64(len(fmt.Sprint(f.name)))
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
	if s.name == "/" {
		return os.ModeDir | 0755
	}
	return 0644
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
	// Add entry for root
	// Loop over our map, add an entry if it is positive
	go func() {
		c <- stat{name: "/", file: &fakefile{name: "/"}}
	LoopMap:
		for name, show := range st.show {
			if show {
				select {
				case c <- stat{name: name, file: &fakefile{name: name}}:
				case <-done:
					break LoopMap
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
