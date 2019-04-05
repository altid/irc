package main

import (
	"log"
	"path"

	"github.com/go-irc/irc"
	"github.com/altid/fslib"
)

// BUG(halfwit): Logs are being created for user events such as client quit
// https://github.com/altid/ircfs/issues/4
func handlerFunc(s *server) irc.HandlerFunc {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		switch m.Command {
		case "PRIVMSG":
			parseForCTCP(c, m, s)
			return
		case "QUIT":
			//TODO(halfwit): When smart filtering is implemented
			// we will check the map of names for channels
			// log to that channel when we're connected to it
			// and logging is enabled/smart filter
			// https://github.com/altid/ircfs/issues/5
			//feed(fbuffer, m.Prefix.Name, s, m)
		case "PART", "KICK", "JOIN", "NICK":
			//name := "server"
			//if c.FromChannel(m) {
			//	name = m.Params[0]
			//}
			//feed(fbuffer, name, s, m)
		case "PING", "PING ZNC":
			c.Writef("PONG %s", m.Params[0])
		case "001":
			s.j <- s.buffs
		case "301":
			feed(fbuffer, m.Params[0], s, m)
		case "333": //topicwhotime <client> <channel> <nick> <setat> unix time
			timeSetAt(s, m)
			return
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
			s.j <- m.Params[1]
			workdir := path.Join(*mtpt, *srv)
			input, err := fslib.NewInput(s, workdir, m.Params[1])
			if err != nil {
				log.Println(err)
				return
			}
			go input.Start()
			if m.Command == "332" {
				title(m.Params[1], s, m)
			}
			// We have to manually send the input event
			s.e <- path.Join(workdir, m.Params[1], "input")
		default:			
			feed(fserver, "server", s, m)
		}
	})
}
