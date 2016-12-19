package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thoj/go-ircevent"
	"github.com/vaughan0/go-ini"
)

type Conf struct {
	Server    string
	UseTLS    string
	Nick      string
	User      string
	Channels  string
	Directory string
}

func writeFile(c *Conf, e *irc.Event) {
	path := filepath.Join(c.Directory, e.Arguments[0])
	if e.Arguments[0] == c.User {
		path = filepath.Join(c.Directory, e.Nick)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Err %s", err)
	}
	defer f.Close()
	f.WriteString(e.Raw)
	f.WriteString("\n")
}

func setupServer(conf ini.File, section string) *irc.Connection {

	c := new(Conf)
	c.Server, _ = conf.Get(section, "Server")
	c.UseTLS, _ = conf.Get(section, "UseTLS")
	c.Nick, _ = conf.Get(section, "Nick")
	c.User, _ = conf.Get(section, "User")
	c.Channels, _ = conf.Get(section, "Channels")
	c.Directory, _ = conf.Get(section, "Directory")

	err := os.MkdirAll(c.Directory, 0744)

	irccon := irc.IRC(c.Nick, c.User)
	//TODO: Flag out
	irccon.Debug = true
	irccon.VerboseCallbackHandler = true

	if c.UseTLS == "true" {
		irccon.UseTLS = true
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range strings.Split(c.Channels, ", ") {
			irccon.Join(channel)
		}
	})
	irccon.AddCallback("PRIVMSG", func(e *irc.Event) {
		writeFile(c, e)
	})
	irccon.AddCallback("CTCP_ACTION", func(e *irc.Event) {
		writeFile(c, e)
	})
	irccon.AddCallback("TOPIC", func(e *irc.Event) {
		writeFile(c, e)
	})
	irccon.AddCallback("366", func(e *irc.Event) {})
	err = irccon.Connect(c.Server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return nil
	}
	go irccon.Loop()
	return irccon

}
