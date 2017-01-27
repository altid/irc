package main

import (
	"log"
	"os"
	"path"

	"aqwari.net/net/styx"
)

type client struct {
	buffer  string
	server  string
	status  string
	sidebar string
}

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
	c := new(client)
	//c.buffer = st.buffer
	//c.server = st.server
	c.buffer = "#ubqt"
	c.server = "freenode"
	c.status = "This is a status, deal bro.\n"
	c.sidebar = "some\ncrap\nthat\nis\nfun\n"
	for s.Next() {
		t := s.Request()
		name := path.Base(t.Path())
		fi := &stat{name: name, file: &fakefile{name: name, event: st.event}}
		p := path.Join(*inPath, c.server, c.buffer)
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(fi, nil)
		case styx.Topen:
			switch name {
			case "/":
				t.Ropen(mkdir(st), nil)
			case "main":
				t.Ropen(os.Open(p))
			default:
				t.Ropen(fi, nil)
			}
		case styx.Tstat:
			switch name {
			case "main":
				t.Rstat(os.Stat(p))
			default:
				t.Rstat(fi, nil)
			}
		case styx.Tcreate:
			t.Rerror("permission denied")
		case styx.Tremove:
			t.Rerror("permission denied")

		}
	}
}
