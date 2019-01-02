package main

import (
	"bufio"
	"log"
	"net"
	"path"
	"text/template"
	"strings"
	"github.com/go-irc/irc"
)

type Servers struct {
	servers []*Server
}

func GetServers(confs []*Config) *Servers {
	var servlist []*Server
	for _, conf := range confs {
		conn, err := GetConn(conf)
		if err != nil {
			log.Printf("Unable to connect: %s\n", conn)
			continue
		}
		userconf := irc.ClientConfig{
			User: conf.User,
			Nick: conf.Nick,
			Name: conf.Name,
			Pass: conf.Pass,
		}

		server := &Server{
			conf: userconf,
			theme: conf.Theme,
			buffers: conf.Chans,
			addr: conf.Addr,
			conn: conn,
			filter: conf.Filter,
			log: conf.Log,
			fmt: conf.Fmt,
			exit: make(chan struct{}),
		}
		server.conf.Handler = NewHandlerFunc(server)
		servlist = append(servlist, server)	
	}
	return &Servers{servers: servlist}
}

// Run - Attempt to start all servers
func (s *Servers) Run() {
	for _, server := range s.servers {
		client := irc.NewClient(server.conn, server.conf)
		go client.Run()
	}
	// Hacky, but should do the trick
	refCount := len(s.servers)
	for {
		for i := 0; i < refCount; i++ {
			select {
			case <-s.servers[i].exit:
				refCount--
			}
		}
		if refCount < 1 {
			break
		}
	}
}

type Server struct {
	conn net.Conn
	conf irc.ClientConfig
	addr string
	theme string
	buffers string
	filter string
	log string
	fmt map[string]*template.Template
	exit chan struct{}
}

func (s *Server) parseControl(r *Reader, c *irc.Client) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			log.Printf("Command unsupported: %s\n", line)
			return
		}
		nick := c.CurrentNick()
		msg := parseControlLine(nick, line)
		switch msg.Command {
		case "QUIT":
			close(s.exit)
			s.conn.Close()
			break
		case "JOIN":
			c.WriteMessage(msg)
			srvdir := path.Join(*base, s.addr)
			logdir := path.Join(s.log, s.addr)
			err := CreateChannel(msg.Params[0], srvdir, logdir)
			if err != nil {
				log.Print(err)
			}
			// Request data from the channel on connection
			// TODO: We may need other data to fill out our files
			//topic := newCTCP(nick, "TOPIC", msg.Params[0])
			//c.WriteMessage(topic)
			mode := newCTCP(nick, "MODE", msg.Params[0])
			c.WriteMessage(mode)
			list := newCTCP(nick, "LIST", msg.Params[0])
			c.WriteMessage(list)
		default:
			c.WriteMessage(msg)
		}
	}
}

func (s *Server) parseInput(current string, r *Reader, c *irc.Client) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// Create and send a message
		nick := c.CurrentNick()
		msg := &irc.Message{
			Prefix: &irc.Prefix{Name: nick},
			Params: []string{current, line},
			Command: "PRIVMSG",
		}
		c.WriteMessage(msg)
		writeTo(path.Join(current, "feed"), nick, s, msg, SelfMsg)
	}
}

func newCTCP(nick, command, target string) *irc.Message {
	message := &irc.Message{
		Prefix: &irc.Prefix{Name: nick},
		Params: []string{target},
		Command: command,
	}
	return message
}

// TODO: Handle a metric ton more interesting messages + ctcp
func parseControlLine(nick, line string) *irc.Message {
	token := strings.Fields(line)
	message := &irc.Message{
		Prefix: &irc.Prefix{Name: nick},
		Params: token[1:],
	}
	switch token[0] {
	case "msg", "m":
		message.Command = "PRIVMSG"
	case "join",  "j":
		message.Command = "JOIN"
	case "part", "p":
		message.Command = "PART"
	case "quit", "q":
		message.Command = "QUIT"
	default:
		message.Command = "NONE"
	}
	return message
}
