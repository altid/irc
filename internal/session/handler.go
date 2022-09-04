package session

import (
	"time"

	"github.com/altid/libs/service/commander"
	"gopkg.in/irc.v3"
)

// BUG(halfwit): Logs are being created for user events such as client quit
// https://github.com/altid/ircfs/issues/4
func handlerFunc(s *Session) irc.HandlerFunc {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		switch m.Command {
		case "PRIVMSG":
			parseForCTCP(c, m, s)
			return
		case "QUIT":
			//TODO(halfwit): When smart filteringf is implemented
			// we will check the map of names for channels
			// log to that channel when we're connected to it
			// and logging is enabled/smart filter
			// https://github.com/altid/ircfs/issues/5
			feed(fbuffer, m.Prefix.Name, s, m)
		case "PART", "KICK", "JOIN", "NICK":
			name := "server"
			if c.FromChannel(m) {
				name = m.Params[0]
			}
			feed(fbuffer, name, s, m)
		case "PING", "PING ZNC":
			c.Writef("PONG %s", m.Params[0])
		case "001":
			if s.conf.Nick != "" {
				c.Writef("NICK %s", s.conf.Nick)
			}
			for _, buff := range getChans(s.Defaults.Buffs) {
				s.j <- &commander.Command{
					Name: "open",
					Args: []string{buff},
				}
			}
			m.Params[0] = "server"
			m.Params[1] = "connected"
			status(s, m)
		case "301":
			feed(fbuffer, m.Params[0], s, m)
		case "333": //topicwhotime <client> <channel> <nick> <setat> unix time
			timeSetAt(s, m)
			return
		case "MODE", "324":
			//m.Params[0] = "server"
			//m.Params[1] = "mode"
			//status(s, m)
		//case "305": //BACK
		//case "306": //AWAY
		// Sidebar
		//case "353": list of names
		//<client> <symbol> <channel> :[prefix]<nick>{ [prefix]<nick>}
		//case "366": // End of names
		//<client> <channel>
		// Title
		case "TOPIC":
			feed(fbuffer, m.Params[0], s, m)
			title(m.Params[1], s, m)
		case "331", "332":
			// Make sure we start listener and add tab
			for _, buff := range getChans(m.Params[1]) {
				s.j <- &commander.Command{
					Name: "open",
					Args: []string{buff},
				}
			}
			if m.Command == "332" {
				// Give the join time to propagate
				// TODO(halfwit) Create the directory for title if none exists
				time.AfterFunc(time.Second*2, func() { title(m.Params[1], s, m) })
			}
		default:
			feed(fserver, "server", s, m)
		}
	})
}