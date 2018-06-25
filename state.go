package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"strconv"
	"strings"
	"text/template"

	"github.com/lrstanley/girc"
	"github.com/vaughan0/go-ini"
)

type State struct {
	sync.Mutex
	irc		map[string]*girc.Client
	tablist map[string]string
	done    chan error
	chanFmt *template.Template
	selfFmt*template.Template
	ntfyFmt *template.Template
	servFmt *template.Template
	highFmt *template.Template
	actiFmt *template.Template
	modeFmt *template.Template
}

func (st *State) parseFormat(conf ini.File) {
	//Set some pretty printed defaults
	chanFmt := `[#5F87A7]({{.Name}}) {{.Data}}`
	selfFmt := `[#076678]({{.Name}}) {{.Data}}`
	highFmt := `[#9d0007]({{.Name}}) {{.Data}}`
	ntfyFmt := `[#5F87A7]({{.Name}}) {{.Data}}`
	servFmt := `--[#5F87A7]({{.Name}}) {{.Data}}--`
	actiFmt := `[#5F87A7( \* {{.Name}}) {{.Data}}`
	modeFmt := `--[#787878](Mode [{{.Data}}] by {{.Name}})`
	for key, value := range conf["options"] {
		switch key {
		case "channelfmt":
			chanFmt = value
		case "notificationfmt":
			ntfyFmt = value
		case "highfmt":
			highFmt = value
		case "selffmt":
			selfFmt = value
		case "actifmt":
			actiFmt = value
		case "modefmt":
			modeFmt = value
		}
	}
	st.chanFmt = template.Must(template.New("chan").Parse(chanFmt))
	st.ntfyFmt = template.Must(template.New("ntfy").Parse(ntfyFmt))
	st.servFmt = template.Must(template.New("serv").Parse(servFmt))
	st.selfFmt = template.Must(template.New("self").Parse(selfFmt))
	st.highFmt = template.Must(template.New("high").Parse(highFmt))
	st.actiFmt = template.Must(template.New("acti").Parse(actiFmt))
	st.modeFmt = template.Must(template.New("mode").Parse(modeFmt))
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

func (st *State) Initialize(chanlist []string, conf *girc.Config, section string) error {
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
	// See if JOIN events get shown for our clients as well
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
		return err
	}
	os.MkdirAll(path.Join(chanpath, "server"), 0777)
	client.Connect()
	return nil
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
	st.parseFormat(conf)
	
	var ircConf *girc.Config
	for section := range conf {
		switch section {
		case "options":
			continue
		default: 
			ircConf = st.parseOptions(conf, section)
			chanlist := st.parseChannels(conf, section)
			err = st.Initialize(chanlist, ircConf, section)		
		}
	}
	return err
}
