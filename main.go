package main

import (
	"fmt"
	"log"

	"github.com/thoj/go-ircevent"
	"aqwari.net/net/styx"
	"github.com/vaughan0/go-ini"
)

/* Make sure we associate our connections with directiories */
type ircState struct {
	irccon irc.Connection;
	dir string
}

func main() {
//TODO: set up the filesystem, before we connect
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon := make([]irc.Connection, 1)
	for section, _ := range conf {
		if section == "options" {
			//setupFiles(conf, section)
			continue
		}
		irccon = append(irccon, *setupServer(conf, section))
	}

	h := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			//TODO: Somewhere here, we will send of the irc message
			switch t := s.Request().(type) {
			case styx.Tstat:
				t.Rstat(directory{}, nil)
			case styx.Twalk:
				t.Rwalk(directory{}, nil)
			case styx.Topen:
				t.Ropen(directory{}, nil)
			}
		}
	})

	//TODO: log.Fatal(styx.ListenAndServe(port, h))
	log.Fatal(styx.ListenAndServe(":564", h))
}
