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
	Name      string
}

func writeFile(c *Conf, e *irc.Event) {
	p := filepath.Join(*inPath, c.Name, e.Arguments[0])
	if e.Arguments[0] == c.User {
		p = filepath.Join(*inPath, c.Name, e.Nick)
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Err %s", err)
	}
	defer f.Close()
	f.WriteString(e.Raw)
	f.WriteString("\n")
}

//TODO: Server messages need to go to named file, ie: ~/freenode/freenode

func setupServer(conf ini.File, section string) *irc.Connection {
	var ok bool
	c := new(Conf)
	c.Server, ok = conf.Get(section, "Server")
	if !ok {
		fmt.Printf("Server entry missing in %s", section)
		return nil
	}
	c.UseTLS, ok = conf.Get(section, "UseTLS")
	if !ok {
		fmt.Printf("nonfatal: UseTLS entry missing in %s", section)
	}
	c.Nick, ok = conf.Get(section, "Nick")
	if !ok {
		fmt.Printf("Nick entry missing in %s", section)
		return nil
	}
	c.User, ok = conf.Get(section, "User")
	if !ok {
		fmt.Printf("nonfatal: User entry missing in %s", section)
	}
	c.Channels, ok = conf.Get(section, "Channels")
	if !ok {
		fmt.Printf("nonfatal: Channels section missing in %s", section)
	}
	c.Name, _ = conf.Get(section, "Name")
	if !ok {
		fmt.Printf("Name entry missing in %s", section)
		return nil
	}

	err := os.MkdirAll(filepath.Join(*inPath, c.Name), 0744)
	if err != nil {
		fmt.Printf("Err %s", err)
		return nil
	}

	irccon := irc.IRC(c.Nick, c.User)
	irccon.Debug = *debug
	irccon.VerboseCallbackHandler = *verbose

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
