package main

import (
	"log"
	"os"
	"path"

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

func newStat(file string) *stat {
	s := &stat{name: file, file: &fakefile{v: 
}

// Serve9P - Called on client connection
func (st *State) Serve9P(s *styx.Session) {
	//TODO: Serve up some initial data for our connection on read
	for s.Next() {
		t := s.Request
		fi := newStat(path.Base(t.Path()))

	}
}
