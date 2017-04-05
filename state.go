package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/lrstanley/girc"
	"github.com/ubqt-systems/ubqtlib"
	"github.com/vaughan0/go-ini"
)

func parseOptions(srv *ubqtlib.Srv, conf ini.File) {
	for key, value := range conf["options"] {
		if value == "show" {
			srv.AddFile(key)
		}
	}
}

// initialize - Read config and set up IRC sessions per entry
// we also log to a filesystem, and set up defaults
func (st *State) initialize(srv *ubqtlib.Srv) error {
	//st.ctl = getCtl()
	conf, err := ini.LoadFile(*conf)
	if err != nil {
		return err
	}
	parseOptions(srv, conf)
	srv.AddFile("ctl")
	srv.AddFile("feed")
	for section := range conf {
		if section == "options" {
			continue
		}
		server, ok := conf.Get(section, "Server")
		if !ok {
			fmt.Println("server entry not found")
		}
		p, ok := conf.Get(section, "Port")
		port, _ := strconv.Atoi(p)
		if !ok {
			fmt.Println("No port set, using 6667")
			port = 6667
		}
		nick, ok := conf.Get(section, "Nick")
		if !ok {
			fmt.Println("nick entry not found")
		}
		user, ok := conf.Get(section, "User")
		if !ok {
			fmt.Println("user entry not found")
		}
		name, ok := conf.Get(section, "Name")
		if !ok {
			fmt.Println("name entry not found")
		}
		channels, _ := conf.Get(section, "Channels")
		chanlist := strings.Split(channels, ",")
		ircConf := girc.Config{
			Server: server,
			Port:   port,
			Nick:   nick,
			User:   user,
			Name:   name,
		}
		client := girc.New(ircConf)
		client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
			for _, channel := range chanlist {
				if strings.Contains(channel, " ") {
					// We have a password
					channel := strings.Fields(channel)
					c.Commands.JoinKey(channel[0], channel[1])
				} else {
					c.Commands.Join(channel)
				}
			}
		})
		client.Handlers.Add(girc.ADMIN, st.writeServer)
		client.Handlers.Add(girc.AWAY, st.writeChannel)
		client.Handlers.Add(girc.INFO, st.writeServer)
		client.Handlers.Add(girc.INVITE, st.writeServer)
		client.Handlers.Add(girc.ISON, st.writeServer)
		client.Handlers.Add(girc.KICK, st.writeChannel)
		client.Handlers.Add(girc.KILL, st.writeChannel)
		client.Handlers.Add(girc.LIST, st.writeServer)
		client.Handlers.Add(girc.LUSERS, st.writeServer)
		client.Handlers.Add(girc.MODE, st.writeChannel)
		client.Handlers.Add(girc.MOTD, st.writeServer)
		client.Handlers.Add(girc.NAMES, st.writeChannel)
		client.Handlers.Add(girc.NICK, st.writeChannel)
		client.Handlers.Add(girc.NOTICE, st.writeChannel)
		client.Handlers.Add(girc.OPER, st.writeServer)
		client.Handlers.Add(girc.PRIVMSG, st.writeChannel)
		client.Handlers.Add(girc.SERVER, st.writeServer)
		client.Handlers.Add(girc.SERVICE, st.writeServer)
		client.Handlers.Add(girc.SERVLIST, st.writeServer)
		client.Handlers.Add(girc.SQUERY, st.writeChannel)
		client.Handlers.Add(girc.STATS, st.writeServer)
		client.Handlers.Add(girc.SUMMON, st.writeServer)
		client.Handlers.Add(girc.TIME, st.writeServer)
		client.Handlers.Add(girc.TOPIC, st.writeChannel)
		client.Handlers.Add(girc.USER, st.writeServer)
		client.Handlers.Add(girc.USERHOST, st.writeServer)
		client.Handlers.Add(girc.USERS, st.writeServer)
		client.Handlers.Add(girc.VERSION, st.writeServer)
		client.Handlers.Add(girc.WALLOPS, st.writeServer)
		client.Handlers.Add(girc.WHO, st.writeServer)
		client.Handlers.Add(girc.WHOIS, st.writeServer)
		client.Handlers.Add(girc.WHOWAS, st.writeServer)
		err = client.Connect()
		if err != nil {
			log.Fatalf("an error occured while attempting to connect to %s: %s", client.Server(), err)
			return err
		}
		// This is a bit odd, as we reassign this for every server.
		st.irc["default"] = client
		st.irc[server] = client
		//TODO: If we have a password, scrub it out here
		st.clients["default"] = &Client{server: server, channel: chanlist[0]}
		// Fire off IRC connection
		go client.Loop()
	}
	return nil
}
