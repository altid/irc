package main

import (
	"io"
	"os"
	"time"

	"aqwari.net/net/styx"
	"github.com/vaughan0/go-ini"
)
//TODO: Modify this to also hold a styx type
type Directory struct {
	dir     string
	title   string
	status  string
	sidebar string
}

func setupFiles(conf ini.File, section string) *Directory {
	d := new(Directory)
	d.dir, _ = conf.Get(section, "Directory")
	d.title, _ = conf.Get(section, "Title")
	d.status, _ = conf.Get(section, "Status")
	d.sidebar, _ = conf.Get(section, "Sidebar")
	return d
}

func (Directory) Readdir(n int) ([]os.FileInfo, error) { return nil, io.EOF }
func (Directory) Mode() os.FileMode                    { return os.ModeDir | 0777 }
func (Directory) IsDir() bool                          { return true }
func (Directory) ModTime() time.Time                   { return time.Now() }
func (Directory) Name() string                         { return "" }
func (Directory) Size() int64                          { return 0 }
func (Directory) Sys() interface{}                     { return nil }

func (d Directory) Serve9P(s *styx.Session) {
	for s.Next() {
		t := s.Request()
		switch t := t.(type) {
		case styx.Twalk:
			//TODO: Have this return our directory
			// if notitle, etc
			t.Rwalk(d, nil)
		case styx.Topen:
			t.Ropen(d, nil)
		case styx.Tstat:
			t.Rstat(d, nil)
		}
	}
}
