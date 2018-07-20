package main

// TODO: Move all file creation/directory to appropriate files
// Read on an event loop to augment state where necessary  

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"strconv"
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

func (st *State) parseOptions(conf ini.File, section string) (*girc.Config) {
	server, ok := conf.Get(section, "Server")
	if ! ok {
		log.Println("Server entry not found!")
	}
	p, ok := conf.Get(section, "Port")
	port, _ := strconv.Atoi(p)
	if !ok {
		fmt.Println("No port set, using default")
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
	return &girc.Config{Server: server, Port: port, Nick: nick, User: user, Name: name, ServerPass: pw}
}

func (st *State) parseChannels(conf ini.File, section string) []string {
	channels, _ := conf.Get(section, "Channels")
	return strings.Split(channels, ",")
}

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

func newState() *State {
	irc := make(map[string]*girc.Client)
	tab := make(map[string]string)
	done := make(chan error)
	return &State{irc: irc, tablist: tab, done: done}
}

// initialize - Read config and set up IRC sessions per entry
func (st *State) OutLoop() error {
	conf, err := ini.LoadFile(*conf)
	if err != nil {
		return err
	}
	st.cfg = ParseFormat(conf)
	
	var ircConf *girc.Config
	for section := range conf {
		switch section {
		case "options":
			continue
		default: 
			ircConf = st.parseOptions(conf, section)
			chanlist := st.parseChannels(conf, section)
			go st.Initialize(chanlist, ircConf, section)		
		}
	}
	return err
}
