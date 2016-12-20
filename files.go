package main

import (
	"errors"
	"io"
	"path"
	"os"
	"time"

	"aqwari.net/net/styx"
	//"github.com/vaughan0/go-ini"
)

type fakefile struct {
	buf string
	name string
	offset int64
}

func (f *fakefile) ReadAt(p []byte, off int64) (int, error) {
	//markdown, f.name
	n := copy (p, f.buf)
	return n, nil
}

func (f *fakefile) WriteAt(p []byte, off int64) (int, error) {
	if off != f.offset {
		return 0, errors.New("no seeking")
	}
	n := 0
	f.offset += int64(n)
	return 0, nil
}

func (f *fakefile) Close() error {
	//set notitle, for example
	return nil
}

func (f *fakefile) size() int64 {
	return int64(len(f.buf))
}

type stat struct {
	name     string
	file *fakefile
}
func (s *stat) Name() string { return s.name }
func (s *stat) Sys() interface{} { return s.file }
func (s *stat) ModTime() time.Time { return time.Now() }
func (s *stat) IsDir() bool { return s.Mode().IsDir() }
func (s *stat) Mode() os.FileMode {
	if s.file.name != s.name {
		return os.ModeDir | 0755
	}
	return 0644
}
func (s *stat) Size() int64 { return s.file.size() }

type Directory struct {
	name string
	title   string
	status  string
	sidebar string
	tabs string
	main string
	c chan stat
	done chan struct{}
}

func (d *Directory) Readdir(n int) ([]os.FileInfo, error) {
	var err error
	fi := make([]os.FileInfo,0, 10)
	for i:= 0; i < n; i++ {
		s, ok := <-d.c
		if !ok {
			err = io.EOF
			break
		}
		fi = append(fi, &s)
	}
	return fi, err
}

func (d *Directory) Close() error {
	close(d.done)
	return nil
}


//TODO: Will have to use this interface I guess.
func (d *Directory) Serve9P(s *styx.Session) {
	for s.Next() {
		t := s.Request()
		fi := &stat{name: path.Base(t.Path()), file: &fakefile{}}
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(fi, nil)
		case styx.Topen:
			t.Ropen(d.c, nil)
		case styx.Tstat:
			t.Rstat(fi, nil)
		}
	}
}
