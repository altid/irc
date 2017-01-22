package main

import (
	"github.com/lionkov/go9p/p"
	"github.com/lionkov/go9p/p/srv"
)

// Write - append entry to input history, fire off message
func (i *Input) Write(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	i.history = append(i.history[:], buf...)
	i.ch <- buf
	return len(buf), nil
}

// Wstat - simple stat value returned
func (i *Input) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil

}

// Remove - We may want to hide input, for whatever reason
func (i *Input) Remove(fid *srv.FFid) error {
	i.show = false
	return nil

}

func (i *Input) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	c := copy(buf, i.history[offset:])
	return c, nil
}

func (c *Current) Read(fid *srv.FFid, buff []byte, offset uint64) (int, error) {
	//Open file for reading
	return 0, nil
}

//Write
//Read
//Wstate
//Remove
