package session 

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

	"github.com/altid/ircfs/internal/format"
	"github.com/altid/libs/service"
	"github.com/altid/libs/config/types"
	"github.com/altid/libs/markup"
	"gopkg.in/irc.v3"
)

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

type Session struct {
	cancel   context.CancelFunc
	conn     net.Conn
	conf     irc.ClientConfig
	e        chan string // events
	i        chan string // inputs
	j        chan string // joins
	m        chan *msg   // messages
	Defaults *Defaults
    Verbose  bool
	debug    func(ctlItem, ...interface{})
}

type Defaults struct { 
	Address string       `altid:"address,prompt:IP Address of IRC server you wish to connect to"`
	SSL     string       `altid:"ssl,prompt:SSL mode,pick:none|simple|certificate"`
	Port    int          `altid:"port,no_prompt"`
	Auth    types.Auth   `altid:"auth,Authentication method to use:"`
	Filter  string       `altid:"filter,no_prompt"`
	Nick    string       `altid:"nick,prompt:Enter your IRC nickname (this is what will be shown on messages you send)"`
	User    string       `altid:"user,no_prompt"`
	Name    string       `altid:"name,no_prompt"`
	Buffs   string       `altid:"buffs,no_prompt"`
	Logdir  types.Logdir `altid:"logdir,no_prompt"`
	TLSCert string       `altid:"tlscert,no_prompt"`
	TLSKey  string       `altid:"tlskey,no_prompt"`
}

func (s *Session) Parse() {
	s.m = make(chan *msg)
	s.e = make(chan string)
	s.j = make(chan string)
	s.i = make(chan string)
	s.debug = func(ctlItem, ...interface{}) {}

	s.conf = irc.ClientConfig{
		User:    s.Defaults.User,
		Nick:    s.Defaults.Nick,
		Name:    s.Defaults.Name,
		Pass:    string(s.Defaults.Auth),
		Handler: handlerFunc(s),
	}

	if s.Verbose {
		s.debug = ctlLogging
	}
}

func (s *Session) Run(c *service.Control, cmd *service.Command) error {
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
		if e := c.CreateBuffer(cmd.Args[0]); e != nil {
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
	case "close":
		// IRC buffers do not allow spaces
		s.debug(ctlPart, cmd.Args[0])
		if e := c.DeleteBuffer(cmd.Args[0]); e != nil {
			return e
		}

		_, err := fmt.Fprintf(s.conn, "PART %s\n", cmd.Args[0])
		s.debug(ctlSucceed, "part")
		return err
	case "open":
		if e := c.CreateBuffer(cmd.Args[0]); e != nil {
			return e
		}

		s.debug(ctlJoin, cmd.Args[0])
		if cmd.Args[0][0] == '#' {
			if _, e := fmt.Fprintf(s.conn, "JOIN %s\n", cmd.Args[0]); e != nil {
				return e
			}
		}

		s.i <- cmd.Args[0]
		s.e <- path.Join(cmd.Args[0], "input")

		s.debug(ctlSucceed, "join")
		return nil
	default:
		return fmt.Errorf("unsupported command %s", cmd.Name)
	}

	s.debug(ctlSucceed, cmd)
	return nil
}

func (s *Session) Quit() {
	s.cancel()
}

// input is always sent down raw to the server
func (s *Session) Handle(bufname string, l *markup.Lexer) error {
	data, err := format.Input(l)
	if err != nil {
		return err
	}

	s.debug(ctlInput, bufname, data)
	if _, e := fmt.Fprintf(s.conn, ":%s PRIVMSG %s :%s\n", s.conf.Name, bufname, data); e != nil {
		return e
	}

        s.m <- &msg{
		data: data,
		from: s.conf.Nick,
		buff: bufname,
	}

	s.debug(ctlSucceed, "input")
	return nil
}

func (s *Session) Start(ctx context.Context, c *service.Control) error {
	if e := s.connect(ctx); e != nil {
		return e
	}
	s.debug(ctlStart, s.Defaults.Address, s.Defaults.Port)
	go s.fileListener(ctx, c)
	return nil
}

// Tie the utility functions like title and feed to the fileWriter
func (s *Session) fileListener(ctx context.Context, c *service.Control) {
	for {
		select {
		case e := <-s.e:
			s.debug(ctlEvent, e)
			//c.Event(e)
		case j := <-s.j:
			buffs := getChans(j)
			for _, buff := range buffs {
				if !c.HasBuffer(buff) {
					cmd := &service.Command{
						Name: "open",
						Args: []string{buff},
					}

					go func() {
						s.Run(c, cmd)
						//c.Input(buff)
					}()
				}
			}
		case m := <-s.m:
			if e := fileWriter(c, m); e != nil {
				errorWriter(c, e)
			}
		//case b := <-s.i:
			//if e := c.Input(b); e != nil {
			//	errorWriter(c, e)
			//}
		case <-ctx.Done():
			return
		}
	}

}

func (s *Session) connect(ctx context.Context) error {
	var tlsConfig *tls.Config

	s.debug(ctlStart, s.Defaults.Address, s.Defaults.Port)
	dialString := fmt.Sprintf("%s:%d", s.Defaults.Address, s.Defaults.Port)
	dialer := &net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", dialString)
	if err != nil {
		return err
	}

	switch s.Defaults.SSL {
	case "none":
		s.conn = conn
		s.debug(ctlRun)
		return nil
	case "simple":
		tlsConfig = &tls.Config{
			ServerName:         dialString,
			InsecureSkipVerify: true,
		}
	case "certificate":
		cert, err := tls.LoadX509KeyPair(s.Defaults.TLSCert, s.Defaults.TLSKey)
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
		m := args[0].(*service.Command)
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
