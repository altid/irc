package main

import(
	"fmt"
	"time"

	"github.com/go-irc/irc"
)

func InitHandler(channels string, server string, format *Format) (irc.Handler) {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		switch m.Command {
		// This is sent on server connection, join channels here
		case "001":
			c.Writef("JOIN %s\n", channels) 
			return
		case "NOTICE":
			WriteToFile(&Msg{Name: m.Prefix.Name, Data: m.Trailing()}, server, "feed", format.ntfyFmt)
		//case "PRIVMSG":
		// - :ACTION
		// - :TOPIC
		// - :FINGER
		// - etcetera
		//if from channel
		//WriteToFile(channel feed)
		//if from user
		//initdirectory
		//WriteToFile(user feed)
		//case "JOIN":
		// if the user is us, initdirectory
		// JOIN for our user implies we're joining a channel. We need to clear out sidebar so we can harvest the name list without a FSM
		//case "PART":
			//WriteToFile(channel feed)
		//case "KICK"
			//WriteToFile(channel feed)
		//case "MODE"
			//WriteToFile(channel status)
			//WriteToFile(channel feed)
		case "TIME":
			t := time.Now()
			c.Write(fmt.Sprintf("TIME %s\n", t.Format("14:33:14 19-Mar-2010")))
		//case "TOPIC"
			//WriteToFile(channel title)
			//WriteToFile(channel feed)
		case "VERSION":
			c.Write("ubqt-ircfs v0.0.0")
		//case "FINGER"
			//WriteToFile(server feed)
		//case "USERINFO"
			//WriteToFile(server feed)
		//case "CLIENTINFO"
			//WriteToFile(server feed)
		case "SOURCE":
			c.Write("https://github.com/ubqt-systems/ircfs")
		//case "301" // <client> <nick> :<message> //away message reply
		//case "305" // no longer away
		//case "306" // now away
		//case "332" // topic - log to channel, as well as set up title
		//case "333" // who set the topic, when - log to channel
		//case "375" // Start of message of the day
		//case "372" // MOTD
		//case "376" // End of MOTD, MODE
		//case "353" // List of names - set up sidebar
		//case "366" // End of name list
		case "PING":
			c.Writef("PONG %s", m.Params[0])
		//case "QUIT"
		default: // Log to server for all other messages so far
			WriteToFile(&Msg{Name: m.Prefix.String(), Data: m.Trailing()}, server, "feed", format.servFmt)
		}
	})
}
