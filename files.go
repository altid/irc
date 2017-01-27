package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

type fakefile struct {
	st     *State
	c      *client
	name   string
	offset int64
}

func (f *fakefile) ReadAt(p []byte, off int64) (int, error) {
	var c string
	switch f.name {
	case "input":
		c = string(f.st.input[:])
	case "ctl":
		c = ":fake\n:list\n:of\n:crap\n"
		//c := f.st.getCtl()
	case "sidebar":
		//c := f.st.getNicks(f.c)
		c = "fake\nsidebar\njohn\nbob\npeter\npaul\n"
	case "status":
		//c := f.st.getStatus(f.c)
		c = "Mock status\n"
	case "tabs":
		//c := f.st.getTabs()
		c = "banana cream pie\n"
	case "title":
		c = "irc"
	default:
		return 0, nil
	}
	n := copy(p, c[off:])
	return n, nil
}

// Called on input and ctl
func (f *fakefile) WriteAt(p []byte, off int64) (int, error) {
	if f.name == "input" {
		f.st.input = append(f.st.input[off:], p...)
	}
	//f.st.event <- string(p[off:])
	return 0, nil
}

func (f *fakefile) Close() error {
	return nil
}

func (f *fakefile) size() int64 {
	switch f.name {
	case "/":
		return 0
	case "input":
		return int64(len(f.st.input))
		//case "ctl":
		//	return int64(len(f.st.getCtl()))
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
	return time.Now()
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
				case c <- stat{name: name, file: &fakefile{name: name, st: st}}:
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
