package main

import (
	"github.com/lionkov/go9p/p"
	"github.com/lionkov/go9p/p/srv"
)

// Write - simply send on channel to parser
func (i *Input) Write(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	i.ch <- buf
	var err error
	return 0, err
}

// Wstat - simple stat value returned
func (i *Input) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil

}

// Remove - We may want to hide input, for whatever reason
func (i *Input) Remove(fid *srv.FFid) error {
	return nil

}

func (i *Input) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	c := copy(buf, "Read")
	return c, nil
}

//Write
//Read
//Wstate
//Remove
