package main

import(
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/go-irc/irc"
)
// TODO: Just log to user/chan/server all messages based on params[0] = currentnick, 
// is.FromChannel(). Good enough.
// Privmsg is a mess, it is what it is. 
func privmsg(srv *Server, c *irc.Client, m *irc.Message) *Data {
	switch { //params: <target>,target <text>
	case m.Params[0] == c.CurrentNick(): // dm - requires further filtering
		switch m.Params[1] {
		// These whitespaces are magic, don't touch them!
		case "VERSION":
		case "FINGER":
			return NewData(m.Prefix.Name, m.Params[1], srv.addr, "server", "feed")
		default:
			return NewData(m.Prefix.Name, m.Params[1], srv.addr, path.Join("messages", m.Prefix.Name), "feed")
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

func joinchannels(c *irc.Client, srv *Server) {
	for _, channel := range srv.channels {
		// TODO: Sanitize our lists
		c.Writef("JOIN %s\n", channel)
	}
}

	
func (srv *Server)InitHandlers(format *Format) (irc.Handler) {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		switch m.Command {
		case "001":
			joinchannels(c, srv)
		case "PRIVMSG":
			writeToFile(privmsg(srv, c, m), format.chanFmt)
		case "ACTION":
			writeToFile(privmsg(srv, c, m), format.actiFmt)
 		case "PART":
		case "QUIT":
		case "KICK":
		case "JOIN":
			// If this is our user joining, we want to clean up nicknames list
			// Set up input
			//if ! srv.joinpartquit {
				//WriteToFile(NewData(m.Prefix.Name, m.Params[1], srv.addr, path.Join("channels", strings.TrimLeft(m.Params[0], "#")), "feed"), format.chanFmt)
			return
		case "NOTICE": // <target> <text>
			writeToFile(NewData(m.Prefix.Name, m.Params[1], srv.addr, "server", "feed"), format.ntfyFmt)

		case "MODE": // <target> [<modestring>[<mode arguments>]]
			writeToFile(modemsg(srv, c, m), format.modeFmt)
		case "TIME":
			t := time.Now()
			c.Write(fmt.Sprintf("TIME %s\n", t.Format("14:33:14 19-Mar-2010")))
		case "TOPIC": // <channel> <topic>
			filePath := path.Join("channels", strings.TrimLeft(m.Params[0], "#"))
			message := "has changed the topic to \"" + m.Params[1] + "\""
			setTopic(srv.addr, filePath, m.Params[1])
			writeToFile(NewData(m.Prefix.Name, message, srv.addr, filePath, "feed"), format.chanFmt)
		case "VERSION":
			c.Write("ubqt-ircfs 0.0.0 https://github.com/ubqt-systems/ircfs")
		case "USERINFO":
			c.Write("USERINFO %s\n", srv.conf.Name)
		case "CLIENTINFO":
			c.Write("CLIENTINFO PING SOURCE TIME USERINFO VERSION")
		//case "301": // <client> <nick> :<message> //away message reply
		//case "305": // no longer away
		//case "306": // 
		case "332": // <client> <channel> :<topic>
			filePath := path.Join("channels", strings.TrimLeft(m.Params[1], "#"))
			message := "topic - " + m.Params[2]
			setTopic(srv.addr, filePath, m.Params[2])
			writeToFile(NewData(m.Prefix.Name, message, srv.addr, filePath, "feed"), format.chanFmt)
		case "333": // <client> <channel> <nick> <setat>
			filePath := path.Join("channels", strings.TrimLeft(m.Params[1], "#"))
			message := " set topic at " + m.Params[3]
			writeToFile(NewData(m.Params[2], message, srv.addr, filePath, "feed"), format.chanFmt)
		case "375": // Start of message ofthe day
		case "372": // MOTD
		case "376": // End of MOTD, MODE
			msgToFile(path.Join(srv.addr, "server", "feed"), m.Trailing())
		case "353": // <client> <symbol> <channel> :[prefix]<nick>{ [prefix]<nick>}
			filePath := path.Join(srv.addr, "channels", strings.TrimLeft(m.Params[2], "#"))
			msgToFile(path.Join(filePath, "feed"), m.Params[3])
			msgToFile(path.Join(filePath, "sidebar"), m.Params[3])
		case "366": // <client> <channel> End of names
			filePath := path.Join(srv.addr, "channels", strings.TrimLeft(m.Params[1], "#"), "feed")
			msgToFile(filePath, m.Trailing())
		case "PING":
			c.Writef("PONG %s", m.Params[0])
		case "SOURCE":
			c.Write("https://github.com/ubqt-systems/ircfs")
		}
	})
}

