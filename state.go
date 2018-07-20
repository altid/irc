package main

// TODO: Move all file creation/directory to appropriate files
// Read on an event loop to augment state where necessary  

// TODO: Read through each server entry

import (
	"log"
	"os"
	"path"
	"sync"
	"strings"

	"github.com/lrstanley/girc"
	"github.com/vaughan0/go-ini"
)

type State struct {
	sync.Mutex
	irc		map[string]*girc.Client
	tablist map[string]string
	done    chan error
	cfg *Cfg
}
// TODO: Try to break all of this out of st *State
func (st *State) Initialize(chanlist []string, conf *girc.Config, section string) {
	client := girc.New(*conf)
	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		for _, channel := range chanlist {
			if strings.Contains(channel, " ") {
				// A space in the string implies a password entry
				// TODO: Switch logic to act on number of members in channel[]
				channel := strings.Fields(channel)
				c.Cmd.JoinKey(channel[0], channel[1])
			} else {
				c.Cmd.Join(channel)
			}
			buffer := path.Join(*inPath, c.Config.Server, channel)
			err := os.MkdirAll(buffer, 0777)
			if err != nil {
				log.Print(err)
			}
		}
	})
	// TODO: Update timestamps for named client on all of these, and test whether to show event
	client.Handlers.Add(girc.JOIN, st.join)
	client.Handlers.Add(girc.PART, st.part)
	client.Handlers.Add(girc.QUIT, st.quitServer)
	client.Handlers.Add(girc.AWAY, st.writeFeed)

	// clean and write to server
	client.Handlers.Add(girc.MOTD, st.writeServer)
	client.Handlers.Add(girc.ADMIN, st.writeServer)
	client.Handlers.Add(girc.INFO, st.writeServer)
	client.Handlers.Add(girc.INVITE, st.writeServer)
	client.Handlers.Add(girc.ISON, st.writeServer)
	client.Handlers.Add(girc.KILL, st.writeServer)
	client.Handlers.Add(girc.LIST, st.writeServer)
	client.Handlers.Add(girc.LUSERS, st.writeServer)
	client.Handlers.Add(girc.NAMES, st.writeServer)
	client.Handlers.Add(girc.NICK, st.writeServer)
	client.Handlers.Add(girc.OPER, st.writeServer)
	client.Handlers.Add(girc.SERVER, st.writeServer)
	client.Handlers.Add(girc.SERVICE, st.writeServer)
	client.Handlers.Add(girc.SERVLIST, st.writeServer)
	client.Handlers.Add(girc.SQUERY, st.writeServer)
	client.Handlers.Add(girc.STATS, st.writeServer)
	client.Handlers.Add(girc.SUMMON, st.writeServer)
	client.Handlers.Add(girc.TIME, st.writeServer)
	client.Handlers.Add(girc.USERHOST, st.writeServer)
	client.Handlers.Add(girc.USERS, st.writeServer)
	client.Handlers.Add(girc.VERSION, st.writeServer)
	client.Handlers.Add(girc.WALLOPS, st.writeServer)
	client.Handlers.Add(girc.WHO, st.writeServer)
	client.Handlers.Add(girc.WHOIS, st.writeServer)
	client.Handlers.Add(girc.WHOWAS, st.writeServer)
	client.Handlers.Add(girc.KICK, st.writeFeed)

	// Need extra parsing under these
	client.Handlers.Add(girc.NOTICE, st.writeFeed)
	client.Handlers.Add(girc.PRIVMSG, st.writeFeed)
	client.Handlers.Add(girc.TOPIC, st.topic)
	client.Handlers.Add(girc.MODE, st.mode)

	// Ensure our filepath exists
	chanpath := path.Join(*inPath, conf.Server)
	os.MkdirAll(chanpath, 0777)
	if _, err := os.Stat(chanpath); os.IsNotExist(err) {
		log.Print(err)
	}
	os.MkdirAll(path.Join(chanpath, "server"), 0777)

	// Fire off listeners and add our client to master map
	st.irc[conf.Server] = client
	go client.Connect()
	st.CtlLoop(conf.Server)
}

// initialize - Read config and set up IRC sessions per entry
func OutLoop() {
	conf, err := ini.LoadFile(*conf)
	if err != nil {
		log.Fatal(err)
	}
	
	var ircConf *girc.Config
	for section := range conf {
		if section == "options" {
			continue
		}
		ircConf = ParseServer(conf, section)
		chanlist := ParseChannels(conf, section)
		go Initialize(chanlist, ircConf, section)
	}
}

func newState() *State {
	conf, err := ini.LoadFile(*conf)
	if err != nil {
		log.Fatal(err)
	}
	irc := make(map[string]*girc.Client)
	tab := make(map[string]string)
	done := make(chan error)
	return &State{irc: irc, tablist: tab, done: done, cfg: ParseFormat(conf)}
}


