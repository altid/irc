package main

import (
	"fmt"
	"log"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/go-irc/irc"
)

func NewHandlerFunc(srv *Server) irc.HandlerFunc {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		var fileName string
		var msgType MessageType
		switch m.Command {
		case "PRIVMSG":
			msgType, fileName = parseForCTCP(c, m, srv)
		case "QUIT":
			// TODO: This should be logged to all channels that are applicable
			msgType = ChanMsg
			fileName = path.Join("server", "feed")
			if m.Prefix.Name == c.CurrentNick() {
				log.Println("Here.")
			}
		case "PART", "KICK", "JOIN", "NICK":
			msgType = ChanMsg
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
			reader, err := NewReader(path.Join(*mtpt, srv.addr, "ctrl"))
			if err != nil {
				log.Print(err)
				return
			}
			go srv.parseControl(reader, c)
			msgType = ServerMsg
			fileName = path.Join("server", "feed")
			c.Writef("JOIN %s\n", srv.buffers)
		case "300": //none
			return
		case "301", "333": // Client is away, and time when topic set
			msgType = ChanMsg
			fileName = path.Join(m.Params[0], "feed")
		// Status
		case "MODE", "324":
			msgType, fileName = parseForUserMode(m)
			// These both require helpers to check files for strings, and append to files - add to files.go
			/* TODO: Update status bar
			case "305": //BACK
			case "306": //AWAY
			*/
			/* TODO: Parse list of names, update with join/part on a given channel
			// Sidebar
			//case "353": list of names
				//<client> <symbol> <channel> :[prefix]<nick>{ [prefix]<nick>}
			//case "366": // End of names
				//<client> <channel>
			*/
		// Title
		case "TOPIC":
			fileName = path.Join(m.Params[0], "title")
			Title(fileName, srv, m)
			msgType = ChanMsg
			fileName = path.Join(m.Params[0], "feed")
			fmt.Printf("From TOPIC %s\n", m.String())
		case "332": //TOPIC - log to channel and set contents of title
			fileName = path.Join(m.Params[1], "title")
			Title(fileName, srv, m)
			reader, err := NewReader(path.Join(*mtpt, srv.addr, m.Params[1], "input"))
			if err != nil {
				log.Print(err)
				return
			}
			go srv.parseInput(m.Params[1], reader, c)
			return
		case "331": // NOTOPIC
			reader, err := NewReader(path.Join(*mtpt, srv.addr, m.Params[1], "input"))
			if err != nil {
				log.Print(err)
				return
			}
			go srv.parseInput(m.Params[1], reader, c)
			return
		default:
			msgType = ServerMsg
			fileName = path.Join("server", "feed")
		}
		if msgType == None {
			return
		}
		WriteTo(fileName, m.Prefix.Name, srv, m, msgType)
	})
}

func parseForUserMode(m *irc.Message) (MessageType, string) {
	// Initialise `status`
	return None, ""
}

// Another large case to implement CTCP protocol
func parseForCTCP(c *irc.Client, m *irc.Message, s *Server) (MessageType, string) {
	prefix := &irc.Prefix{Name: c.CurrentNick()}
	command := strings.Split(m.Params[1], " ")
	switch command[0] {
	case "ACTION":
		m.Params[1] = strings.Join(command[1:], " ")
		return ActionMsg, path.Join(m.Params[0], "feed")
	case "CLIENTINFO":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "CLIENTINFO",
			Params:  []string{m.Prefix.Name, "ACTION CLIENTINFO FINGER PING SOURCE TIME USER INFO VERSION"},
		})
		return ServerMsg, path.Join("server", "feed")
	case "DDC":
		return None, ""
	case "FINGER":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "FINGER",
			Params:  []string{m.Prefix.Name, "ircfs 0.0.0"},
		})
		return ServerMsg, path.Join("server", "feed")
	case "PING", "PING":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "PONG",
			Params:  []string{m.Prefix.Name, m.Params[1]},
		})
		return None, ""
	case "SOURCE":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "SOURCE",
			Params:  []string{m.Prefix.Name, "https://github.com/ubqt-systems/ircfs"},
		})
		return ServerMsg, path.Join("server", "feed")
	case "TIME":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "TIME",
			Params:  []string{m.Prefix.Name, time.Now().Format(time.RFC3339)},
		})
		return ServerMsg, path.Join("server", "feed")
	case "VERSION":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "VERSION",
			Params:  []string{m.Prefix.Name, "ircfs 0.0.0"},
		})
		return ServerMsg, path.Join("server", "feed")
	case "USERINFO":
		c.WriteMessage(&irc.Message{
			Prefix:  prefix,
			Command: "USERINFO",
			Params:  []string{m.Prefix.Name, c.CurrentNick()},
		})
		return ServerMsg, path.Join("server", "feed")
	}
	format := ChanMsg
	file := path.Join(m.Prefix.Name, "feed")
	if c.FromChannel(m) {
		file = path.Join(m.Params[0], "feed")
	}
	// Highlight
	if strings.Contains(m.Params[1], c.CurrentNick()) {
		Event(path.Join(*mtpt, s.addr, m.Params[0], "notify"), s)
		WriteTo(path.Join(m.Params[0], "notify"), m.Prefix.Name, s, m, HighMsg)
		file = path.Join(m.Params[0], "feed")
		format = HighMsg
	}
	return format, file
}

func parseForFormat(srv *Server, msgType MessageType) *template.Template {
	switch msgType {
	case SelfMsg:
		return srv.fmt["self"]
	case ServerMsg:
		return srv.fmt["server"]
	case TitleMsg:
		return srv.fmt["title"]
	case StatusMsg:
		return srv.fmt["mode"]
	case HighMsg:
		return srv.fmt["highlight"]
	case ActionMsg:
		return srv.fmt["action"]
	}
	return srv.fmt["channel"]
}

