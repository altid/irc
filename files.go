package main

import (
	"github.com/lionkov/go9p/p"
	"github.com/lionkov/go9p/p/srv"
)

// Write - simply send on channel to parser, append to completion list
func (i *Input) Write(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	i.history = append(i.history[:], buf...)
	writeMsg(i, buf[offset:])
	return len(buf), nil
}

// Wstat - simple stat value returned
func (i *Input) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil

}

// Remove - We may want to hide input, for whatever reason
func (i *Input) Remove(fid *srv.FFid) error {
	//Clunk the file
	return nil

}

func (i *Input) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	c := copy(buf, i.history[offset:])
	return c, nil
}

//Write
//Read
//Wstate
//Remove
