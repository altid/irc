package main

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/go-irc/irc"
)

func privmsg(srv *Server, c *irc.Client, m *irc.Message) *Data {
	switch { //params: <target>,target <text>
	case m.Params[0] == c.CurrentNick(): // dm - requires further filtering
		switch m.Params[1] {
		// These whitespaces are magic, don't touch them!
		case "VERSION":
		case "CLIENTINFO":
		case "USERINFO":
		case "TIME":
		case "SOURCE":
			return NewData(m.Prefix.Name, m.Params[1], srv.addr, "server", "feed")
		default:
			filePath := path.Join("messages", m.Prefix.Name)
			// DMs warrant highlights
			writeToEvent(NewData("", "", srv.addr, filePath, "highlight"))
			return NewData(m.Prefix.Name, m.Params[1], srv.addr, filePath, "feed")
		}
	case c.FromChannel(m): // channel
		filePath := path.Join("channels", strings.TrimLeft(m.Params[0], "#"))
		return NewData(m.Prefix.Name, m.Params[1], srv.addr, filePath, "feed")
	}
	return NewData(m.Prefix.Name, m.Params[1], srv.addr, "server", "feed")
}

func modemsg(srv *Server, c *irc.Client, m *irc.Message) *Data {
	// TODO: Set up status bar here as well
	switch { //params: <target> [<modestring>[<mode arguments>]]
	case m.Params[0] == c.CurrentNick():
		return NewData(m.Prefix.Name, m.Params[1], srv.addr, "server", "feed")
	}
	filePath := path.Join("channels", strings.TrimLeft(m.Params[0], "#"))
	return NewData(m.Prefix.Name, m.Params[1], srv.addr, filePath, "feed")
}

func filter(srv *Server, c *irc.Client, m *irc.Message) bool {
	switch srv.filter {
	case "all":
		return false
	case "smart":
		// TODO: check hotlist
		return false
	case "none":
		return true
	}
	return true
}

func joinchannels(c *irc.Client, srv *Server) {
	for _, channel := range srv.channels {
		// TODO: Sanitize our lists
		c.Writef("JOIN %s\n", channel)
	}
}

func (srv *Server) InitHandlers(format *Format) irc.Handler {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		switch m.Command {
		case "001":
			joinchannels(c, srv)
		case "PRIVMSG":
			data := privmsg(srv, c, m)
			writeToFile(c.CurrentNick(), data, format.chanFmt)
			writeToEvent(data)
		case "ACTION":
			data := privmsg(srv, c, m)
			writeToFile(c.CurrentNick(), data, format.actiFmt)
			writeToEvent(data)
		case "PART":
		case "QUIT":
		case "KICK":
			if filter(srv, c, m) {
				filePath := path.Join("channels", strings.TrimLeft(m.Params[0], "#"))
				data := NewData(m.Prefix.Name, m.Params[1], srv.addr, filePath, "feed")
				writeToFile(c.CurrentNick(), data, format.chanFmt)
				writeToEvent(data)
			}
		case "JOIN":
			// Initialize our directory - add to tabs
			if m.Params[0] == c.CurrentNick() {
				return
			}
			//if filter(srv, c, m) {
			//	filePath := path.Join("channels", strings.TrimLeft(m.Params[0], "#"))
			//	data := NewData(m.Prefix.Name, m.Params[1], srv.addr, filePath, "feed")
			//	writeToFile(c.CurrentNick(), data, format.chanFmt)
			//	writeToEvent(data)
			//}
		case "NOTICE": // <target> <text>
			data := NewData(m.Prefix.Name, m.Params[1], srv.addr, "server", "feed")
			writeToFile(c.CurrentNick(), data, format.ntfyFmt)
			writeToEvent(data)

		case "MODE": // <target> [<modestring>[<mode arguments>]]
			data := modemsg(srv, c, m)
			writeToFile(c.CurrentNick(), data, format.modeFmt)
			writeToEvent(data)
		case "TIME":
			t := time.Now()
			c.Write(fmt.Sprintf("TIME %s\n", t.Format("14:33:14 19-Mar-2010")))
		case "TOPIC": // <channel> <topic>
			filePath := path.Join("channels", strings.TrimLeft(m.Params[0], "#"))
			message := "has changed the topic to \"" + m.Params[1] + "\""
			setTopic(srv.addr, filePath, m.Params[1])
			data := NewData(m.Prefix.Name, message, srv.addr, filePath, "feed")
			writeToFile(c.CurrentNick(), data, format.chanFmt)
			writeToEvent(data)
		case "VERSION":
			c.Write("ubqt-ircfs 0.0.0 https://github.com/ubqt-systems/ircfs")
		case "USERINFO":
			c.Writef("USERINFO %s\n", srv.conf.Name)
		case "CLIENTINFO":
			c.Write("CLIENTINFO PING SOURCE TIME USERINFO VERSION")
		case "301": // <client> <nick> :<message> //away message reply
			if filter(srv, c, m) {
				filePath := path.Join("channels", strings.TrimLeft(m.Params[0], "#"))
				message := "is away"
				if len(m.Params) > 2 { //<client> <nick> <message>w
					message = message + ": \"" + m.Params[2] + "\""
				}
				data := NewData(m.Prefix.Name, message, srv.addr, filePath, "feed")
				writeToFile(c.CurrentNick(), data, format.chanFmt)
				writeToEvent(data)
			}
		//case "305": client back // update status bar
		//case "306": client away // update status bar
		case "332": // <client> <channel> :<topic>
			filePath := path.Join("channels", strings.TrimLeft(m.Params[1], "#"))
			message := "topic - " + m.Params[2]
			setTopic(srv.addr, filePath, m.Params[2])
			data := NewData(m.Prefix.Name, message, srv.addr, filePath, "feed")
			writeToFile(c.CurrentNick(), data, format.chanFmt)
			writeToEvent(data)
		case "333": // <client> <channel> <nick> <setat>
			filePath := path.Join("channels", strings.TrimLeft(m.Params[1], "#"))
			message := " set topic at " + m.Params[3]
			data := NewData(m.Params[2], message, srv.addr, filePath, "feed")
			writeToFile(c.CurrentNick(), data, format.chanFmt)
			writeToEvent(data)
		case "375": // Start of message ofthe day
		case "372": // MOTD
		case "376": // End of MOTD, MODE
			filePath := path.Join(srv.addr, "server", "feed")
			msgToFile(filePath, m.Trailing())
			msgToEvent(filePath)
		case "353": // <client> <symbol> <channel> :[prefix]<nick>{ [prefix]<nick>}
			filePath := path.Join(srv.addr, "channels", strings.TrimLeft(m.Params[2], "#"), "feed")
			msgToFile(filePath, m.Trailing())
			msgToEvent(filePath)
		case "366": // <client> <channel> End of names
			filePath := path.Join(srv.addr, "channels", strings.TrimLeft(m.Params[1], "#"), "feed")
			msgToFile(filePath, m.Trailing())
			msgToEvent(filePath)
		case "PING":
			c.Writef("PONG %s", m.Params[0])
		case "SOURCE":
			c.Write("https://github.com/ubqt-systems/ircfs")
		}
	})
}
