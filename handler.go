package main

import(
	"fmt"
	"github.com/go-irc/irc"
)

func InitHandler(channels string) (irc.Handler) {
	return irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
		switch m.Command {
		// This is sent on server connection, join channels here
		case "001":
			c.Writef("JOIN %s\n", channels) 
		//case "INVITE"
		//case "NOTICE":
		//case "PRIVMSG":
		//case "JOIN":
		//case "PART":
		//case "KICK"
		//case "MODE"
		//case "TOPIC"
		//case "301" // <client> <nick> :<message> //away message reply
		//case "305" // no longer away
		//case "306" // now away
		//case "332" // topic
		//case "333" // who set the topic, when
		//case "372" // MOTD
		//case "375" // Start of message of the day
		//case "376" // En of MOTD, MODE
		//case "353" // List of names
		//case "366" // End of name list
		//case "PING"
		//case "QUIT"
		default: // Log to server for all other messages so far
			fmt.Printf("command %s, %s\n", m.Command, m.String())
		}
	})
}
