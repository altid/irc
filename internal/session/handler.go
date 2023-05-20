package session

import (
	//"log"
	"strings"
	"time"

	"github.com/altid/libs/service/commander"
	irc "gopkg.in/irc.v3"
)

var ctcpMsg ctlItem

// BUG(halfwit): Logs are being created for user events such as client quit
// https://github.com/altid/ircfs/issues/4
func handlerFunc(s *Session) irc.HandlerFunc {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		switch m.Command {
		case "PRIVMSG":
			prefix := &irc.Prefix{
				Name: c.CurrentNick(),
			}

			s.debug(ctcpMsg, m)
			token := strings.Split(m.Params[1], " ")
			switch token[0] {
			case "\x01ACTION":
				m.Params[1] = strings.Join(token[1:], " ")
				fn := faction
				if m.Params[0] == prefix.Name {
					fn = fselfaction
				}
				feed(fn, m.Params[0], s.ctrl, m)
			case "\x01CLIENTINFO":
				c.WriteMessage(&irc.Message{
					Prefix:  prefix,
					Command: "CLIENTINFO",
					Params:  []string{m.Prefix.Name, "ACTION CLIENTINFO FINGER PING SOURCE TIME USER INFO VERSION"},
				})
				feed(fserver, "server", s.ctrl, m)
			case "\x01FINGER":
				c.WriteMessage(&irc.Message{
					Prefix:  prefix,
					Command: "FINGER",
					Params:  []string{m.Prefix.Name, "ircfs 0.1.0"},
				})
				feed(fserver, "server", s.ctrl, m)
			case "\x01PING", "PING":
				c.WriteMessage(&irc.Message{
					Prefix:  prefix,
					Command: "PONG",
					Params:  []string{m.Prefix.Name, m.Params[1]},
				})
			case "\x01SOURCE":
				c.WriteMessage(&irc.Message{
					Prefix:  prefix,
					Command: "SOURCE",
					Params:  []string{m.Prefix.Name, "https://github.com/altid/ircfs"},
				})
				feed(fserver, "server", s.ctrl, m)
			case "\x01TIME":
				c.WriteMessage(&irc.Message{
					Prefix:  prefix,
					Command: "TIME",
					Params:  []string{m.Prefix.Name, time.Now().Format(time.RFC3339)},
				})
				feed(fserver, "server", s.ctrl, m)
			case "\x01VERSION":
				c.WriteMessage(&irc.Message{
					Prefix:  prefix,
					Command: "VERSION",
					Params:  []string{m.Prefix.Name, "ircfs 0.1.0"},
				})
				feed(fserver, "server", s.ctrl, m)
			case "\x01USERINFO":
				c.WriteMessage(&irc.Message{
					Prefix:  prefix,
					Command: "USERINFO",
					Params:  []string{m.Prefix.Name, s.conf.Nick},
				})
				feed(fserver, "server", s.ctrl, m)
			default:
				switch {
				// TODO(halfwit) Would prefer to use hostmask matches here
				// Messages the user writes
				case m.Name == c.CurrentNick():
					c := &commander.Command{
						Name: "open",
						Args: []string{m.Params[0]},
					}
					go s.Run(s.ctrl, c)
					feed(fself, m.Params[0], s.ctrl, m)
				// User is highlighted
				case strings.Contains(m.Params[1], c.CurrentNick()):
					if m.Params[0] == "chanserv" || m.Params[0] == "chanserve" {
						m.Params[0] = "server"
					} else {
						feed(fhighlight, m.Params[0], s.ctrl, m)
					}

					m := &msg{
						fn:   fnotification,
						buff: m.Params[0],
						from: m.Name,
						data: m.Trailing(),
					}
					fileWriter(s.ctrl, m)
				// PM received, make sure the file exists
				case m.Params[0] == c.CurrentNick():
					cmd := &commander.Command{
						Name: "open",
						Args: []string{m.Prefix.Name},
					}
					go s.Run(s.ctrl, cmd)
					feed(fbuffer, m.Prefix.Name, s.ctrl, m)
				// Normal message from a buffer
				case c.FromChannel(m):
					cmd := &commander.Command{
						Name: "open",
						Args: []string{m.Params[0]},
					}
					go s.Run(s.ctrl, cmd)
					feed(fbuffer, m.Params[0], s.ctrl, m)
				default:
					feed(fserver, "server", s.ctrl, m)
				}
			}
		case "QUIT":
			//TODO(halfwit): When smart filteringf is implemented
			// we will check the map of names for channels
			// log to that channel when we're connected to it
			// and logging is enabled/smart filter
			// https://github.com/altid/ircfs/issues/5
			feed(fbuffer, m.Prefix.Name, s.ctrl, m)
		case "PART", "KICK", "JOIN", "NICK":
			name := "server"
			if c.FromChannel(m) {
				name = m.Params[0]
			}
			feed(fbuffer, name, s.ctrl, m)
		case "PING", "PING ZNC":
			c.Writef("PONG %s", m.Params[0])
		case "001":
			if s.conf.Nick != "" {
				c.Writef("NICK %s", s.conf.Nick)
			}
			for _, buff := range getChans(s.Defaults.Buffs) {
				cmd := &commander.Command{
					Name: "open",
					Args: []string{buff},
				}

				go s.Run(s.ctrl, cmd)
			}
			title("server", s.ctrl, m)
		case "301":
			feed(fbuffer, m.Params[0], s.ctrl, m)
		case "333": //topicwhotime <client> <channel> <nick> <setat> unix time
			timeSetAt(s.ctrl, m)
			return
		case "MODE", "324":
			m.Params[0] = "server"
			status(s.ctrl, m)
		//case "305": //BACK
		//case "306": //AWAY
		// Sidebar
		//case "353": list of names
		//<client> <symbol> <channel> :[prefix]<nick>{ [prefix]<nick>}
		//case "366": // End of names
		//<client> <channel>
		// Title
		case "TOPIC":
			feed(fbuffer, m.Params[0], s.ctrl, m)
			title(m.Params[1], s.ctrl, m)
		case "331":
			cmd := &commander.Command{
				Name: "open",
				Args: []string{m.Params[1]},
			}
			s.Run(s.ctrl, cmd)
		case "332":
			cmd := &commander.Command{
				Name: "open",
				Args: []string{m.Params[1]},
			}
			go s.Run(s.ctrl, cmd)
			//title(m.Params[1], s.ctrl, m)
		default:
			feed(fserver, "server", s.ctrl, m)
		}
	})
}
