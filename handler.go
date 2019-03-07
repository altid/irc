package main

import (
	"log"
	"path"

	"github.com/go-irc/irc"
	"github.com/ubqt-systems/fslib"
)

func handlerFunc(s *server) irc.HandlerFunc {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		switch m.Command {
		case "PRIVMSG":
			//fn, buffname := parseForCTCP()
			feed(fbuffer, m.Params[0], s, m)
		case "QUIT":
			feed(fbuffer, m.Params[0], s, m)
		case "PART", "KICK", "JOIN", "NICK":
			name := "server"
			if c.FromChannel(m) {
				name = m.Params[0]
			}
			feed(fbuffer, name, s, m)
		case "PING", "PING ZNC":
			c.Writef("PONG %s", m.Params[0])
		case "001":
			c.Writef("JOIN %s\n", s.buffs)
		case "301", "333":
			feed(fbuffer, m.Params[0], s, m)
		case "MODE", "324":
			status(s, m)
		//case "305": //BACK
		//case "306": //AWAY
		// Sidebar
		//case "353": list of names
			//<client> <symbol> <channel> :[prefix]<nick>{ [prefix]<nick>}
		//case "366": // End of names
			//<client> <channel>
		// Title
		case "TOPIC":
			title(m.Params[1], s, m)
			feed(fbuffer, m.Params[0], s, m)
		// This is the title sent on channel connection
		// We use this to start our input listeners
		case "331", "332":
			workdir := path.Join(*mtpt, s.addr)
			input, err := fslib.NewInput(s, workdir, m.Params[1])
			if err != nil {
				log.Println(err)
				return
			}
			go input.Start()
			if m.Command == "332" {
				title(m.Params[1], s, m)
			}
		default:			
			feed(fserver, "server", s, m)
		}
	})
}
