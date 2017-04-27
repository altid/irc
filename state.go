package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/lrstanley/girc"
	"github.com/ubqt-systems/ubqtlib"
	"github.com/vaughan0/go-ini"
)

func newState() *State {
	client := make(map[string]*Client)
	irc := make(map[string]*girc.Client)
	tab := make(map[string]string)
	event := make(chan []byte)
	return &State{clients: client, irc: irc, tablist: tab, event: event}
}

func parseOptions(srv *ubqtlib.Srv, conf ini.File, st *State) {
	//Set some pretty printed defaults
	chanFmt := `[#5F87A7]({{.Name}}) {{.Data}}`
	selfFmt := `[#076678]({{.Name}}) {{.Data}}`
	highFmt := `[#9d0007]({{.Name}}) {{.Data}}`
	ntfyFmt := `[#5F87A7]({{.Name}}) {{.Data}}`
	servFmt := `--[#5F87A7]({{.Name}}) {{.Data}}--`
	for key, value := range conf["options"] {
		if value == "show" {
			srv.AddFile(key)
		}
		switch key {
		case "channelfmt":
			chanFmt = value
		case "notificationfmt":
			ntfyFmt = value
		case "highfmt":
			highFmt = value
		}
	}
	st.chanFmt = template.Must(template.New("chan").Parse(chanFmt))
	st.ntfyFmt = template.Must(template.New("ntfy").Parse(ntfyFmt))
	st.servFmt = template.Must(template.New("serv").Parse(servFmt))
	st.selfFmt = template.Must(template.New("self").Parse(selfFmt))
	st.highFmt = template.Must(template.New("high").Parse(highFmt))
}

// initialize - Read config and set up IRC sessions per entry
// we also log to a filesystem, and set up defaults
func (st *State) initialize(srv *ubqtlib.Srv) error {
	//st.ctl = getCtl()
	conf, err := ini.LoadFile(*conf)
	if err != nil {
		return err
	}
	parseOptions(srv, conf, st)
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
		pw, ok := conf.Get(section, "Password")
		if !ok {
			fmt.Println("password entry not found")
		}
		channels, _ := conf.Get(section, "Channels")
		chanlist := strings.Split(channels, ",")
		ircConf := girc.Config{
			Server:   server,
			Port:     port,
			Nick:     nick,
			User:     user,
			Name:     name,
			ServerPass: pw,
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
		client.Handlers.Add(girc.ADMIN, st.writeFile)
		client.Handlers.Add(girc.AWAY, st.writeFile)
		client.Handlers.Add(girc.INFO, st.writeFile)
		client.Handlers.Add(girc.INVITE, st.writeFile)
		client.Handlers.Add(girc.ISON, st.writeFile)
		client.Handlers.Add(girc.KICK, st.writeFile)
		client.Handlers.Add(girc.KILL, st.writeFile)
		client.Handlers.Add(girc.LIST, st.writeFile)
		client.Handlers.Add(girc.LUSERS, st.writeFile)
		client.Handlers.Add(girc.MODE, st.writeFile)
		client.Handlers.Add(girc.MOTD, st.writeFile)
		client.Handlers.Add(girc.NAMES, st.writeFile)
		client.Handlers.Add(girc.NICK, st.writeFile)
		client.Handlers.Add(girc.NOTICE, st.writeFile)
		client.Handlers.Add(girc.OPER, st.writeFile)
		client.Handlers.Add(girc.PRIVMSG, st.writeFile)
		client.Handlers.Add(girc.SERVER, st.writeFile)
		client.Handlers.Add(girc.SERVICE, st.writeFile)
		client.Handlers.Add(girc.SERVLIST, st.writeFile)
		client.Handlers.Add(girc.SQUERY, st.writeFile)
		client.Handlers.Add(girc.STATS, st.writeFile)
		client.Handlers.Add(girc.SUMMON, st.writeFile)
		client.Handlers.Add(girc.TIME, st.writeFile)
		client.Handlers.Add(girc.TOPIC, st.writeFile)
		client.Handlers.Add(girc.USER, st.writeFile)
		client.Handlers.Add(girc.USERHOST, st.writeFile)
		client.Handlers.Add(girc.USERS, st.writeFile)
		client.Handlers.Add(girc.VERSION, st.writeFile)
		client.Handlers.Add(girc.WALLOPS, st.writeFile)
		client.Handlers.Add(girc.WHO, st.writeFile)
		client.Handlers.Add(girc.WHOIS, st.writeFile)
		client.Handlers.Add(girc.WHOWAS, st.writeFile)
		err = client.Connect()
		if err != nil {
			log.Fatalf("an error occured while attempting to connect to %s: %s", client.Server(), err)
			return err
		}
		// Make sure our directory exists
		filePath := path.Join(*inPath, server)
		os.MkdirAll(filePath, 0777)
		// This is a bit odd, as we reassign this for every server.
		st.irc["default"] = client
		st.irc[server] = client
		//TODO: If we have a password, scrub it out here
		st.clients["default"] = &Client{server: server, channel: chanlist[0]}
		// Fire off IRC connection
		go client.Connect()
	}
	return nil
}
