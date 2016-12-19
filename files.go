package main

import (
	"io"
	"os"
	"time"

	"github.com/vaughan0/go-ini"
)

type Directory struct{
	port string
	title string
	status string
	sidebar string
}

func (Directory) Readdir(n int) ([]os.FileInfo, error) { return nil, io.EOF }
func (Directory) Mode() os.FileMode                    { return os.ModeDir | 0777 }
func (Directory) IsDir() bool                          { return true }
func (Directory) ModTime() time.Time                   { return time.Now() }
func (Directory) Name() string                         { return "" }
func (Directory) Size() int64                          { return 0 }
func (Directory) Sys() interface{}                     { return nil }

func setupFiles(conf ini.File, section string) *Directory {
	d := new(Directory)
	d.port, _ = conf.Get(section, "Port")
	d.title, _ = conf.Get(section, "Title")
	d.status, _ = conf.Get(section, "Status")
	d.sidebar, _ = conf.Get(section, "Sidebar")
	return d
}
