package main

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/thoj/go-ircevent"
	"github.com/vaughan0/go-ini"
)

func setupServer(conf ini.File, section string) *irc.Connection {

	server, _ := conf.Get(section, "Server")
	usetls, _ := conf.Get(section, "UseTLS")

	nick, ok := conf.Get(section, "Nick")
	if !ok {
		fmt.Printf("Err: no Nick in section %s", server)
		return nil
	}
	user, ok := conf.Get(section, "User")
	if !ok {
		fmt.Printf("Err: no User in section %s", server)
		return nil
	}
	channels, ok := conf.Get(section, "Channels")
	if !ok {
		fmt.Printf("Err no Channels in section %s", server)
		return nil
	}
//TODO: --debug
//irccon.Debug = true
//TODO: -v --verbose
//irccon.VerboseCallbackHandler = true
	irccon := irc.IRC(nick, user)
	if usetls == "true" {
		irccon.UseTLS = true
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range strings.Split(channels, ", ") {
			irccon.Join(channel)
		}
	})
	irccon.AddCallback("PRIVMSG", func(e *irc.Event) {
/*TODO: Call writes to filesystem, etc from here */
//If server dir doesn't exist, create dir
//Create or append to files
//e.Nick is name of personsaying the thing
//Arguments[0] is channel
//Arguments[1] is message
		fmt.Printf("%s %s\n", e.Nick, e.Arguments[1])
	})
	irccon.AddCallback("CTCP_ACTION", func(e *irc.Event) {
		fmt.Printf(" * %s %s\n", e.Nick, e.Arguments[1])
	})
	irccon.AddCallback("366", func(e *irc.Event) {})
	err := irccon.Connect(server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return nil
	}
	go irccon.Loop()
	return irccon

}
