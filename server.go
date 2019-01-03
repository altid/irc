package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

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
			conf:    userconf,
			theme:   conf.Theme,
			buffers: conf.Chans,
			addr:    conf.Addr,
			conn:    conn,
			filter:  conf.Filter,
			log:     conf.Log,
			fmt:     conf.Fmt,
		}
		server.conf.Handler = NewHandlerFunc(server)
		servlist = append(servlist, server)
	}
	return &Servers{servers: servlist}
}

// Run - Attempt to start all servers, clean up after
func (s *Servers) Run() {
	// Context to make sure we clean up everything
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(len(s.servers))
	for _, server := range s.servers {
		go func() {
			server.ctx = ctx
			client := irc.NewClient(server.conn, server.conf)
			client.RunContext(ctx)
			wg.Done()
			// Clean up on server exit
			glob := path.Join(*base, server.addr, "*", "feed")
			feeds, err := filepath.Glob(glob)
			if err != nil {
				log.Print(err)
			}
			for _, feed := range feeds {
				go DeleteChannel(feed)
			}
		}()
	}
	wg.Wait()
}

type Server struct {
	fmt     map[string]*template.Template
	conn    net.Conn
	conf    irc.ClientConfig
	ctx	context.Context
	addr    string
	theme   string
	buffers string
	filter  string
	log     string
}

func (s *Server) Event(filename string) {
	fileName := path.Join(*base, s.addr, "event")
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		log.Print(err)
		return
	}
	f.WriteString(filename + "\n")
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
			// Request data from the channel
			// TODO: We may need other data to fill out our files
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
			Prefix:  &irc.Prefix{Name: nick},
			Params:  []string{current, line},
			Command: "PRIVMSG",
		}
		c.WriteMessage(msg)
		writeTo(path.Join(current, "feed"), nick, s, msg, SelfMsg)
	}
}

func newCTCP(nick, command, target string) *irc.Message {
	message := &irc.Message{
		Prefix:  &irc.Prefix{Name: nick},
		Params:  []string{target},
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
	case "join", "j":
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
