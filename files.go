package main

import (
	"io"
	"os"
	"time"
)

type directory struct{}

func (directory) Readdir(n int) ([]os.FileInfo, error) { return nil, io.EOF }
func (directory) Mode() os.FileMode                    { return os.ModeDir | 0777 }
func (directory) IsDir() bool                          { return true }
func (directory) ModTime() time.Time                   { return time.Now() }
func (directory) Name() string                         { return "" }
func (directory) Size() int64                          { return 0 }
func (directory) Sys() interface{}                     { return nil }
