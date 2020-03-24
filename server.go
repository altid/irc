package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"

	"github.com/altid/libs/config"
	"github.com/altid/libs/fs"
	"github.com/altid/libs/markup"
	"gopkg.in/irc.v3"
)

var workdir = path.Join(*mtpt, *srv)

type ctlItem int

const (
	ctlJoin ctlItem = iota
	ctlPart
	ctlStart
	ctlEvent
	ctlMsg
	ctlInput
	ctlRun
	ctlSucceed
	ctlErr
)

type server struct {
	cancel context.CancelFunc
	conn   net.Conn
	conf   irc.ClientConfig
	e      chan string // events
	i      chan string // inputs
	j      chan string // joins
	m      chan *msg   // messages
	d      *defaults
	debug  func(ctlItem, ...interface{})
}

type defaults struct {
	Address string `IP Address of IRC server you wish to connect to:`
	SSL     string `SSL mode: none|simple|certificate`
	Port    int
	Auth    config.Auth `Select authentication type password|factotum|none`
	Filter  string
	Nick    string `Enter your IRC nickname (this is what will be shown on messages you send):`
	User    string
	Name    string
	Buffs   string
	Logdir  config.Logdir
	TLSCert string
	TLSKey  string
}

func (s *server) parse() {
	s.m = make(chan *msg)
	s.e = make(chan string)
	s.j = make(chan string)
	s.i = make(chan string)
	s.debug = func(ctlItem, ...interface{}) {}

	s.conf = irc.ClientConfig{
		User:    s.d.User,
		Nick:    s.d.Nick,
		Name:    s.d.Name,
		Pass:    string(s.d.Auth),
		Handler: handlerFunc(s),
	}

	if *debug {
		s.debug = ctlLogging
	}
}

func (s *server) Open(c *fs.Control, name string) error {
	if e := c.CreateBuffer(name, "feed"); e != nil {
		return e
	}

	s.debug(ctlJoin, name)
	if name[0] == '#' {
		if _, e := fmt.Fprintf(s.conn, "JOIN %s\n", name); e != nil {
			return e
		}
	}

	s.i <- name
	s.e <- path.Join(workdir, name, "input")

	s.debug(ctlSucceed, "join")
	return nil
}

func (s *server) Close(c *fs.Control, name string) error {
	s.debug(ctlPart, name)
	if e := c.DeleteBuffer(name, "feed"); e != nil {
		return e
	}

	_, err := fmt.Fprintf(s.conn, "PART %s\n", name)
	s.debug(ctlSucceed, "part")
	return err
}

func (s *server) Link(c *fs.Control, from, name string) error {
	err := errors.New("link command not supported, please use open/close")
	s.debug(ctlErr, name, err)

	return err
}

func (s *server) Default(c *fs.Control, cmd *fs.Command) error {
	s.debug(ctlMsg, cmd)
	switch cmd.Name {
	case "a", "act", "action", "me":
		if len(cmd.Args) < 1 {
			return errors.New("no action entered")
		}
		line := strings.Join(cmd.Args[1:], " ")
		if e := action(s, cmd.Args[0], line); e != nil {
			return e
		}
	case "msg", "query":
		if len(cmd.Args) < 1 {
			return errors.New("no user specified")
		}
		if e := c.CreateBuffer(cmd.Args[0], "feed"); e != nil {
			return e
		}

		s.i <- cmd.Args[0]
		if len(cmd.Args) > 1 {
			line := strings.Join(cmd.Args[1:], " ")
			if e := pm(s, line); e != nil {
				return e
			}
		}
	case "nick":
		s.conf.Name = cmd.Args[0]
		fmt.Fprintf(s.conn, "NICK %s\n", cmd.Args[0])
	default:
		return fmt.Errorf("Unknown command %s", cmd.Name)
	}

	s.debug(ctlSucceed, cmd)
	return nil
}

func (s *server) Quit() {
	s.cancel()
}

func (s *server) Restart(c *fs.Control) error {
	return nil
}

func (s *server) Refresh(c *fs.Control) error {
	return nil
}

// input is always sent down raw to the server
func (s *server) Handle(bufname string, l *markup.Lexer) error {
	m, err := input(l)
	if err != nil {
		return err
	}

	s.debug(ctlInput, bufname, m.data)
	if _, e := fmt.Fprintf(s.conn, ":%s PRIVMSG %s :%s\n", s.conf.Name, bufname, m.data); e != nil {
		return e
	}

	m.from = s.conf.Nick
	m.buff = bufname
	s.m <- m

	s.debug(ctlSucceed, "input")
	return nil
}

// Tie the utility functions like title and feed to the fileWriter
func (s *server) fileListener(ctx context.Context, c *fs.Control) {
	for {
		select {
		case e := <-s.e:
			s.debug(ctlEvent, e)
			c.Event(e)
		case j := <-s.j:
			buffs := getChans(j)
			for _, buff := range buffs {
				if !c.HasBuffer(buff, "feed") {
					go s.Open(c, buff)
				}
			}
		case m := <-s.m:
			if e := fileWriter(c, m); e != nil {
				errorWriter(c, e)
			}
		case b := <-s.i:
			in, e := fs.NewInput(s, workdir, b, *debug)
			if e != nil {
				errorWriter(c, e)
				continue
			}
			in.Start()
		case <-ctx.Done():
			return
		}
	}

}

func (s *server) connect(ctx context.Context) error {
	var tlsConfig *tls.Config

	s.debug(ctlStart, s.d.Address, s.d.Port)
	dialString := fmt.Sprintf("%s:%d", s.d.Address, s.d.Port)
	dialer := &net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", dialString)
	if err != nil {
		return err
	}

	switch s.d.SSL {
	case "none":
		s.conn = conn
		return nil
	case "simple":
		tlsConfig = &tls.Config{
			ServerName:         dialString,
			InsecureSkipVerify: true,
		}
	case "certificate":
		cert, err := tls.LoadX509KeyPair(s.d.TLSCert, s.d.TLSKey)
		if err != nil {
			return err
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				cert,
			},
			ServerName: dialString,
		}
	}

	tlsconn := tls.Client(conn, tlsConfig)
	if e := tlsconn.Handshake(); e != nil {
		return e
	}

	s.conn = tlsconn
	s.debug(ctlRun)

	return nil
}

func ctlLogging(ctl ctlItem, args ...interface{}) {
	l := log.New(os.Stdout, "ircfs ", 0)

	switch ctl {
	case ctlSucceed:
		l.Printf("%s succeeded\n", args[0])
	case ctlJoin:
		l.Printf("join target=\"%s\"\n", args[0])
	case ctlStart:
		l.Printf("start addr=\"%s\", port=%d\n", args[0], args[1])
	case ctlRun:
		l.Println("connected")
	case ctlPart:
		l.Printf("part target=\"%s\"\n", args[0])
	case ctlEvent:
		l.Printf("event data=\"%s\"\n", args[0])
	case ctlInput:
		l.Printf("input target=\"%s\" data=\"%s\"\n", args[0], args[1])
	case ctlMsg:
		m := args[0].(*fs.Command)
		line := strings.Join(m.Args, " ")
		l.Printf("%s data=\"%s\"\n", m.Name, line)
	case ctlErr:
		l.Printf("error buffer=\"%s\" err=\"%v\"\n", args[0], args[1])
	// This will be a lot of line noise
	case ctcpMsg:
		m := args[0].(*irc.Message)
		l.Printf("ctcp: name=\"%s\" prefix=\"%s\" params=\"%v\"\n", m.Name, m.Prefix, m.Params)
	}
}
