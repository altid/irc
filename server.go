package main

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/thoj/go-ircevent"
	"github.com/vaughan0/go-ini"
)

func setupServer(conf ini.File, section string) *irc.Connection {

//TODO: Proper error checking here
	nick, _ := conf.Get(section, "Nick")
	user, _ := conf.Get(section, "User")
	server, _ := conf.Get(section, "Server")
	channels, _ := conf.Get(section, "Channels")
	usetls, _ := conf.Get(section, "UseTLS")

	irccon := irc.IRC(nick, user)
	irccon.Debug = true
	if usetls == "true" {
		irccon.UseTLS = true
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range strings.Split(channels, ", ") {
			irccon.Join(channel)
		}
	})
	irccon.AddCallback("366", func(e *irc.Event) {
/*TODO: Call writes to filesystem, etc from here */
	})
	irccon.VerboseCallbackHandler = true
	err := irccon.Connect(server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return nil
	}
	go irccon.Loop()
	return irccon

}
