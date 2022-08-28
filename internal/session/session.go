package session

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/altid/ircfs/internal/format"
	"github.com/altid/libs/config/types"
	"github.com/altid/libs/markup"
	"github.com/altid/libs/service/commander"
	"github.com/altid/libs/service/controller"
	"gopkg.in/irc.v3"
)

type ctlItem int

const (
	ctlJoin ctlItem = iota
	ctlPart
	ctlStart
	ctlEvent
	ctlMsg
	ctlCommand
	ctlInput
	ctlRun
	ctlSucceed
	ctlErr
)

type Session struct {
	Client   *irc.Client
	ctx      context.Context
	cancel   context.CancelFunc
	conn     net.Conn
	conf     irc.ClientConfig
	j        chan *commander.Command // joins
	m        chan *msg               // messages
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
	s.j = make(chan *commander.Command)
	s.debug = func(ctlItem, ...interface{}) {}
	s.ctx, s.cancel = context.WithCancel(context.Background())

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

func (s *Session) Run(c controller.Controller, cmd *commander.Command) error {
	s.debug(ctlMsg, cmd)
	switch cmd.Name {
	case "a", "act", "action", "me":
		if len(cmd.Args) < 1 {
			e := errors.New("no action entered")
			s.debug(ctlErr, e)
			return e
		}
		line := strings.Join(cmd.Args[1:], " ")
		if e := action(s, cmd.Args[0], line); e != nil {
			s.debug(ctlErr, e)
			return e
		}
	case "msg", "query":
		if len(cmd.Args) < 1 {
			e := errors.New("no user specified")
			s.debug(ctlErr, e)
			return e
		}
		if e := c.CreateBuffer(cmd.Args[0]); e != nil {
			s.debug(ctlErr, e)
			return e
		}

		if len(cmd.Args) > 1 {
			line := strings.Join(cmd.Args[1:], " ")
			if e := pm(s, line); e != nil {
				s.debug(ctlErr, e)
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
			s.debug(ctlErr, e)
			return e
		}

		_, err := fmt.Fprintf(s.conn, "PART %s\n", cmd.Args[0])
		s.debug(ctlSucceed, "part")
		if err != nil {
			s.debug(ctlErr, err)
		}
		return err
	case "open":
		if e := c.CreateBuffer(cmd.Args[0]); e != nil {
			s.debug(ctlErr, e)
			return e
		}

		s.debug(ctlJoin, cmd.Args[0])
		if cmd.Args[0][0] == '#' {
			if _, e := fmt.Fprintf(s.conn, "JOIN %s\n", cmd.Args[0]); e != nil {
				s.debug(ctlErr, e)
				return e
			}
		}

		s.debug(ctlSucceed, "join")
		return nil
	default:
		e := fmt.Errorf("unsupported command %s", cmd.Name)
		s.debug(ctlErr, e)
	}

	s.debug(ctlSucceed, cmd)
	return nil
}

func (s *Session) Quit() {
	s.cancel()
}

// input is always sent down raw to the server
func (s *Session) Handle(bufname string, l *markup.Lexer) error {
	s.debug(ctlInput, l)
	data, err := format.Input(l)
	if err != nil {
		s.debug(ctlErr, err)
		return err
	}

	if _, e := fmt.Fprintf(s.conn, ":%s PRIVMSG %s :%s\n", s.conf.Name, bufname, data); e != nil {
		s.debug(ctlErr, e)
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

func (s *Session) Start(c controller.Controller) error {
	go s.fileListener(s.ctx, c)
	if err := s.connect(s.ctx); err != nil {
		s.debug(ctlErr, err)
		return err
	}

	s.Client = irc.NewClient(s.conn, s.conf)
	return s.Client.Run()
}

func (s *Session) Listen(c controller.Controller) {
	err := make(chan error)
	go func(err chan error) {
		err <- s.Start(c)
	}(err)

	select {
	case e := <-err:
		s.debug(ctlErr, e)
		log.Fatal(e)
	case <-s.ctx.Done():
	}
}

func (s *Session) Command(cmd *commander.Command) error {
	go func(s *Session, cmd *commander.Command) {
		s.j <- cmd
	}(s, cmd)

	return nil
}

// Tie the utility functions like title and feed to the fileWriter
func (s *Session) fileListener(ctx context.Context, c controller.Controller) {
	for {
		select {
		case j := <-s.j:
			s.debug(ctlCommand, j)
			if e := s.Run(c, j); e != nil {
				s.debug(ctlErr, e)
				errorWriter(c, e)
			}
		case m := <-s.m:
			if e := fileWriter(c, m); e != nil {
				s.debug(ctlErr, e)
				errorWriter(c, e)
			}
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
		s.debug(ctlErr, err)
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
			s.debug(ctlErr, err)
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
		s.debug(ctlErr, e)
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
		if m, ok := args[0].(*markup.Lexer); ok {
			data, _ := m.Bytes()
			l.Printf("input data=\"%s\"", data)
		}
	case ctlCommand:
		m := args[0].(*commander.Command)
		l.Printf("command name=\"%s\" heading=\"%d\" sender=\"%s\" args=\"%s\" from=\"%s\"", m.Name, m.Heading, m.Sender, m.Args, m.From)
	case ctlMsg:
		m := args[0].(*commander.Command)
		line := strings.Join(m.Args, " ")
		l.Printf("%s data=\"%s\"\n", m.Name, line)
	case ctlErr:
		l.Printf("error err=\"%v\"\n", args[0])
	// This will be a lot of line noise
	case ctcpMsg:
		m := args[0].(*irc.Message)
		l.Printf("ctcp: name=\"%s\" prefix=\"%s\" params=\"%v\"\n", m.Name, m.Prefix, m.Params)
	}
}
