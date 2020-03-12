package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/altid/libs/config"
	"github.com/altid/libs/fs"
	"github.com/altid/libs/markup"
	"github.com/go-irc/irc"
)

var workdir = path.Join(*mtpt, *srv)

type server struct {
	conn   net.Conn
	conf   irc.ClientConfig
	cert   tls.Certificate
	e      chan string // events
	j      chan string // joins
	m      chan *msg   // messages
	i      chan string // inputs
	done   chan struct{}
	addr   string
	buffs  string
	filter string
	log    string
	port   string
	ssl    string
}

func (s *server) parse(c *config.Config) {
	s.m = make(chan *msg)
	s.e = make(chan string)
	s.j = make(chan string)
	s.i = make(chan string)
	s.cert, _ = c.SSLCert()
	s.addr, _ = c.Search("address")
	s.buffs, _ = c.Search("buffers")
	s.filter, _ = c.Search("filter")
	s.log = c.Log()
	s.port, _ = c.Search("port")
	s.ssl, _ = c.Search("ssl")
	pass, _ := c.Password()

	s.conf = irc.ClientConfig{
		User:    c.MustSearch("user"),
		Nick:    c.MustSearch("nick"),
		Name:    c.MustSearch("name"),
		Pass:    pass,
		Handler: handlerFunc(s),
	}
}

func (s *server) Open(c *fs.Control, name string) error {
	if e := c.CreateBuffer(name, "feed"); e != nil {
		fmt.Println(e)
		return e
	}

	if name[0] == '#' {
		if _, e := fmt.Fprintf(s.conn, "JOIN %s\n", name); e != nil {
			return e
		}
	}

	s.i <- name
	defer c.Event(path.Join(workdir, name, "input"))

	return nil
}

func (s *server) Close(c *fs.Control, name string) error {
	if e := c.DeleteBuffer(name, "feed"); e != nil {
		return e
	}

	_, err := fmt.Fprintf(s.conn, "PART %s\n", name)
	return err
}

func (s *server) Link(c *fs.Control, from, name string) error {
	return errors.New("link command not supported, please use open/close")
}

func (s *server) Default(c *fs.Control, cmd, from, m string) error {
	switch cmd {
	case "a", "act", "action", "me":
		return action(s, from, m)
	case "msg", "query":
		// we don't want to send a JOIN message, so we don't simply s.j <- t[0]
		t := strings.Fields(m)

		if e := c.CreateBuffer(t[0], "feed"); e != nil {
			return e
		}

		s.i <- t[0]
		return pm(s, m)
	case "nick":
		s.conf.Name = m
		fmt.Fprintf(s.conn, "NICK %s\n", m)

		return nil
	}

	return fmt.Errorf("Unknown command %s", cmd)
}

// input is always sent down raw to the server
func (s *server) Handle(bufname string, l *markup.Lexer) error {
	m, err := input(l)
	if err != nil {

		return err
	}

	if _, e := fmt.Fprintf(s.conn, ":%s PRIVMSG %s :%s\n", s.conf.Name, bufname, m.data); e != nil {
		return e
	}

	m.from = s.conf.Nick
	m.buff = bufname
	s.m <- m

	return nil
}

// Tie the utility functions like title and feed to the fileWriter
func (s *server) fileListener(ctx context.Context, c *fs.Control) {
	for {
		select {
		case e := <-s.e:
			c.Event(e)
		case j := <-s.j:
			buffs := getChans(j)
			for _, buff := range buffs {
				if !c.HasBuffer(buff, "feed") {
					s.Open(c, buff)
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
			} else {
				in.Start()
			}
		case <-ctx.Done():
			s.conn.Close()
			return
		}
	}

}

func (s *server) connect(ctx context.Context) error {
	var tlsConfig *tls.Config

	dialString := s.addr + ":" + s.port
	dialer := &net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", dialString)
	if err != nil {
		return err
	}

	switch s.ssl {
	case "simple":
		tlsConfig = &tls.Config{
			ServerName:         dialString,
			InsecureSkipVerify: true,
		}
	case "certificate":
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				s.cert,
			},
			ServerName: dialString,
		}

	default:
		s.conn = conn
		return nil
	}

	tlsconn := tls.Client(conn, tlsConfig)
	if e := tlsconn.Handshake(); e != nil {
		return e
	}

	s.conn = tlsconn
	return nil
}
