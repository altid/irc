package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"aqwari.net/net/styx"
)

// Run - Fires off a goroutine per connection
func (st *State) Run() error {
	var srv styx.Server
	if *verbose {
		srv.ErrorLog = log.New(os.Stderr, "", 0)
	}
	if *debug {
		srv.TraceLog = log.New(os.Stderr, "", 0)
	}
	srv.Addr = *addr
	srv.Handler = st
	// Long running, will return error if one occurs
	err := srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

// Serve9P - Called on client connection
func (st *State) Serve9P(s *styx.Session) {
	//TODO: Serve up some initial data for our connection on read
	for s.Next() {
		t := s.Request()
		name := path.Base(t.Path())
		fi := &stat{name: name, file: &fakefile{name: name}}
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(fi, nil)
		case styx.Topen:
			switch name {
			case "/":
				t.Ropen(mkdir(st), nil)
			default:
				t.Ropen(strings.NewReader(fmt.Sprint(st)), nil)
			}
		case styx.Tstat:
			t.Rstat(fi, nil)
		case styx.Tcreate:
			t.Rerror("permission denied")
		case styx.Tremove:
			t.Rerror("permission denied")

		}
	}
}
