package main

import (
	"fmt"
	"log"

	"github.com/thoj/go-ircevent"
	"aqwari.net/net/styx"
	"github.com/vaughan0/go-ini"
)

func main() {
	d := new(Directory)
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon := make([]irc.Connection, 1)
	for section, _ := range conf {
		if section == "options" {
			d = setupFiles(conf, section)
			continue
		}
		irccon = append(irccon, *setupServer(conf, section))
	}
	h := styx.HandlerFunc(func(s *styx.Session) {
		for s.Next() {
			//TODO: Somewhere here, we will send of the irc message
			switch t := s.Request().(type) {
			case styx.Tstat:
				t.Rstat(Directory{}, nil)
			case styx.Twalk:
				t.Rwalk(Directory{}, nil)
			case styx.Topen:
				t.Ropen(Directory{}, nil)
			}
		}
	})

	log.Fatal(styx.ListenAndServe(d.port, h))
}
