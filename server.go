package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/go-irc/irc"
	"github.com/ubqt-systems/fslib"
)

type server struct {
	conn   net.Conn
	conf   irc.ClientConfig
	cert   tls.Certificate
	m      chan *msg
	done   chan struct{}
	addr   string
	buffs  string
	filter string
	log    string
	port   string
	ssl    string
}

func newServer(c *config) *server {
	m := make(chan *msg)
	s := &server{
		m:      m,
		addr:   c.addr,
		buffs:  c.chans,
		cert:   c.cert,
		filter: c.filter,
		log:    c.log,
		port:   c.port,
		ssl:    c.ssl,
	}
	conf := irc.ClientConfig{
		User:    c.user,
		Nick:    c.nick,
		Name:    c.name,
		Pass:    c.pass,
		Handler: handlerFunc(s),
	}
	s.conf = conf
	return s
}

func (s *server) Open(c *fslib.Control, name string) error {
	_, err := fmt.Fprintf(s.conn, "JOIN %s\n", name)
	return err
}

func (s *server) Close(c *fslib.Control, name string) error {
	_, err := fmt.Fprintf(s.conn, "PART %s\n", name)
	return err
}

func (s *server) Default(c *fslib.Control, msg string) error {
	token := strings.Fields(msg)
	switch token[0] {
	case "join": 
		return s.Open(c, strings.Join(token[1:], " "))
	case "part":
		return s.Close(c, strings.Join(token[1:], " "))
	case "msg", "query": // util.go
		return pm(s, strings.Join(token[1:], " "))
	case "nick":
		// Make sure we update s.conf.Name when we update username
		s.conf.Name = token[1]
		fmt.Fprintf(s.conn, "NICK %s\n", token[1])
		return nil
	}
	return fmt.Errorf("Unknown command %s", token[0])
}

// input is always sent down raw to the server
func (s *server) Handle(bufname, msg string) error {
	buffer := path.Base(bufname)
	_, err := fmt.Fprintf(s.conn, ":%s PRIVMSG %s :%s\n", s.conf.Name, buffer, msg)
	return err
}

// Tie the utility functions like title and feed to the fileWriter
func (s *server) fileListener(ctx context.Context, c *fslib.Control) {
	for {
		select {
		case m := <- s.m:
			fileWriter(c, m)
		case <- ctx.Done():			
			s.conn.Close()
			return
		}
	}

}

func (s *server) run(conn net.Conn, ctx context.Context) error {
	client := irc.NewClient(conn, s.conf)
	return client.RunContext(ctx)
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
			ServerName:   dialString,
		}

	default:
		s.conn = conn
		return nil
	}
	tlsconn := tls.Client(conn, tlsConfig)
	tlsconn.Handshake()
	s.conn = tlsconn
	return nil
}
