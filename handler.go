package main

import (
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/go-irc/irc"
	"github.com/ubqt-systems/fslib"
)


// Make sure we update s.conf.Name when we update username
// Most all action happens here, everything else generally writes to s.conn, which will end up sending messages here. This allows us to not have to poll for content to be available
func handlerFunc(s *server) irc.HandlerFunc {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		var fileName string
		var msgType messageType
		switch m.Command {
		case "PRIVMSG":
			msgType, fileName = parseForCTCP(c, m, s)
		case "QUIT":
			// TODO: This should be logged to all channels that are applicable
			msgType = chanMsg
			fileName = path.Join("server", "feed")
		case "PART", "KICK", "JOIN", "NICK":
			msgType = chanMsg
			switch {
			case m.Prefix.Name == c.CurrentNick():
				fileName = path.Join("server", "feed")
			case c.FromChannel(m):
				fileName = path.Join(m.Params[0], "feed")
			default:
				fileName = path.Join("server", "feed")
			}
		case "PING", "PING ZNC": // we hide PING/PONG
			c.Writef("PONG %s", m.Params[0])
			return
		case "001": // Successfully connected to server
			c.Writef("JOIN %s\n", s.buffs)
			return
		case "300": //none
			return
		case "301", "333": // Client is away, and time when topic set
			msgType = chanMsg
			fileName = path.Join(m.Params[0], "feed")
		// Status
		case "MODE", "324":
			msgType, fileName = parseForUserMode(m)
		// TODO: Update status bar
		case "305": //BACK
		case "306": //AWAY
		// TODO: Parse list of names, update with join/part on a given channel
		// Sidebar
		//case "353": list of names
			//<client> <symbol> <channel> :[prefix]<nick>{ [prefix]<nick>}
		//case "366": // End of names
			//<client> <channel>
		// Title
		case "TOPIC":
			//Title(fileName, srv, m)
			msgType = titleMsg
			// Log title change to the channel
			fileName = path.Join(m.Params[0], "feed")
		case "332": //Title set, this is sent on connection
			workdir := path.Join(*mtpt, s.addr)
			input, err := fslib.NewInput(s, workdir, m.Params[1])
			if err != nil {
				log.Println(err)
				return
			}
			go func() {
				err := input.Start()
				if err != nil {
					log.Print(err)
				}
			}()
			//Title(fileName, srv, m)
			return
		case "331": // NOTOPIC
			workdir := path.Join(*mtpt, s.addr)
			input, err := fslib.NewInput(s, workdir, m.Params[1])
			if err != nil {
				log.Println(err)
				return
			}
			go func() {
				err := input.Start()
				if err != nil {
					log.Print(err)
				}
			}()
			return
		default:
			msgType = serverMsg
			fileName = path.Join("server", "feed")
		}
		if msgType == none {
			return
		}
		// <-&msg{msgType: msgType, fileName: fileName, m: m}
		if fileName == "banana" {
			fmt.Printf("%s %d\n", fileName, msgType)
		}
	})
}

// TODO: Use a nice bitmask for usermode, and stringify the mode based on that
func parseForUserMode(m *irc.Message) (messageType, string) {
	// Initialise `status`
	return none, ""
}

// Another large case to implement CTCP protocol
func parseForCTCP(c *irc.Client, m *irc.Message, s *server) (messageType, string) {
	prefix := &irc.Prefix{Name: c.CurrentNick()}
	command := strings.Split(m.Params[1], " ")
	switch command[0] {
	case "ACTION":
		m.Params[1] = strings.Join(command[1:], " ")
		return actionMsg, path.Join(m.Params[0], "feed")
	case "CLIENTINFO":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "CLIENTINFO",
			Params:  []string{m.Prefix.Name, "ACTION CLIENTINFO FINGER PING SOURCE TIME USER INFO VERSION"},
		})
		return serverMsg, path.Join("server", "feed")
	case "DDC":
		return none, ""
	case "FINGER":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "FINGER",
			Params:  []string{m.Prefix.Name, "ircfs 0.0.0"},
		})
		return serverMsg, path.Join("server", "feed")
	case "PING", "PING":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "PONG",
			Params:  []string{m.Prefix.Name, m.Params[1]},
		})
		return none, ""
	case "SOURCE":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "SOURCE",
			Params:  []string{m.Prefix.Name, "https://github.com/ubqt-systems/ircfs"},
		})
		return serverMsg, path.Join("server", "feed")
	case "TIME":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "TIME",
			Params:  []string{m.Prefix.Name, time.Now().Format(time.RFC3339)},
		})
		return serverMsg, path.Join("server", "feed")
	case "VERSION":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "VERSION",
			Params:  []string{m.Prefix.Name, "ircfs 0.0.0"},
		})
		return serverMsg, path.Join("server", "feed")
	case "USERINFO":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "USERINFO",
			Params:  []string{m.Prefix.Name, c.CurrentNick()},
		})
		return serverMsg, path.Join("server", "feed")
	}
	format := chanMsg
	file := path.Join(m.Prefix.Name, "feed")
	if c.FromChannel(m) {
		file = path.Join(m.Params[0], "feed")
	}
	// Highlight // [halfwit](red) A message for you sir
	//if strings.Contains(m.Params[1], c.CurrentNick()) {
	//fslib.Notification(file, m.Trailing())
	//}
	return format, file
}
