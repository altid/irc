package main

import (
	"fmt"
	"io"
	"os"
	//"path"

	"github.com/lionkov/go9p/p"
	"github.com/lionkov/go9p/p/srv"
)

// Write - append entry to input history, fire off message
func (i *Input) Write(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	i.history = append(i.history[:], buf...)
	//if /buffer | /quit, send it off to ctl
	m := &Message{buf: buf, id: fid.Fid.Fconn.Id}
	i.ch <- m
	return len(buf), nil
}

// Remove - We may want to hide input, for whatever reason
func (i *Input) Remove(fid *srv.FFid) error {
	i.show = false
	return nil

}

// Wstat - So we can stat
func (i *Input) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil
}

func (i *Input) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	c := copy(buf, i.history[offset:])
	return c, nil
}

// Wstat - for current
func (c *Current) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil
}

func (c *Current) Read(fid *srv.FFid, buff []byte, offset uint64) (int, error) {
	//conn := fid.Fid.Fconn.Id
	//p := path.Join(*inPath, c.server[conn], c.buffer[conn])
	file, err := os.Open("/home/halfwit/local/run/irc/freenode/#ubqt")
	if err != nil {
		fmt.Printf("Err %s", err)
	}
	defer file.Close()
	//<-c.ch
	n, err := file.ReadAt(buff, int64(offset))
	if err != nil && err != io.EOF {
		fmt.Printf("Err %s", err)
	}
	return n, nil
}

func (s *Status) Read(fid *srv.FFid, buff []byte, offset uint64) (int, error) {

	b := []byte("hello")
	c := copy(buff, b[offset:])
	return c, nil
}

func (c *Ctl) Read(fid *srv.FFid, buff []byte, offset uint64) (int, error) {
	return 0, nil
}

func (t *Title) Read(fid *srv.FFid, buff []byte, offset uint64) (int, error) {
	b := []byte("#MY PROGRAM")
	c := copy(buff, b[offset:])
	return c, nil
}

func (t *Tabs) Read(fid *srv.FFid, buff []byte, offset uint64) (int, error) {
	return 0, nil
}
