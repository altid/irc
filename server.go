package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/go-irc/irc"
	"github.com/altid/fslib"
	"github.com/altid/cleanmark"
)

type server struct {
	conn   net.Conn
	conf   irc.ClientConfig
	cert   tls.Certificate
	e      chan string // events
        j      chan string // joins
	m      chan *msg   // messages
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
	e := make(chan string)
	j := make(chan string)
	s := &server{
		e:	e,
		m:      m,
		j:	j,
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
	err := c.CreateBuffer(name, "feed")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.conn, "JOIN %s\n", name)
	return err
}

func (s *server) Close(c *fslib.Control, name string) error {
	err := c.DeleteBuffer(name, "feed")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.conn, "PART %s\n", name)
	return err
}

func (s *server) Link(c *fslib.Control, from, name string) error {
	return fmt.Errorf("link command not supported, please use open/close\n")
}

func (s *server) Default(c *fslib.Control, cmd, from, msg string) error {
	switch cmd {
	case "a", "act", "action", "me":
		return action(s, from, msg)
	case "msg", "query":
		return pm(s, msg)
	case "nick":
		// Make sure we update s.conf.Name when we update username
		s.conf.Name = msg
		fmt.Fprintf(s.conn, "NICK %s\n", msg)
		return nil
	}
	return fmt.Errorf("Unknown command %s", cmd)
}

// input is always sent down raw to the server
func (s *server) Handle(bufname string, l *cleanmark.Lexer) error {
	var m strings.Builder
	buffer := path.Base(bufname)
	for {
		i := l.Next()
		switch i.ItemType {
		case cleanmark.EOF:
			_, err := fmt.Fprintf(s.conn, ":%s PRIVMSG %s :%s\n", s.conf.Name, buffer, m.String())
			s.m <- &msg{
				buff: buffer,
				from: s.conf.Nick,
				data: m.String(),
				fn:   fself,
			}
			return err
		case cleanmark.UrlLink, cleanmark.UrlText, cleanmark.ImagePath, cleanmark.ImageLink, cleanmark.ImageText:
			continue
		case cleanmark.ColorText:
			fmt.Println("We made it into color")
			text := i.Data
			i := l.Next()
			if i.ItemType == cleanmark.EOF || i.ItemType != cleanmark.ColorCode {
				return fmt.Errorf("Improperly formatted color code")
			}
			m.WriteString("")
			switch string(i.Data) {
			case cleanmark.White:
				m.WriteString("0")
			case cleanmark.Black:
				m.WriteString("1")
			case cleanmark.Blue:
				m.WriteString("2")
			case cleanmark.Green:
				m.WriteString("3")
			case cleanmark.Red:
				m.WriteString("4")
			case cleanmark.Brown:
				m.WriteString("5")
			case cleanmark.Purple:
				m.WriteString("6")
			case cleanmark.Orange:
				m.WriteString("7")
			case cleanmark.Yellow:
				m.WriteString("8")
			case cleanmark.LightGreen:
				m.WriteString("9")
			case cleanmark.Cyan:
				m.WriteString("10")
			case cleanmark.LightCyan:
				m.WriteString("11")
			case cleanmark.LightBlue:
				m.WriteString("12")
			case cleanmark.Pink:
				m.WriteString("13")
			case cleanmark.Grey:
				m.WriteString("14")
			case cleanmark.LightGrey:
				m.WriteString("15")
			}
			m.Write(text)
			m.WriteString("")
		case cleanmark.BoldText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case cleanmark.EmphasisText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case cleanmark.UnderlineText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		default:
			m.Write(i.Data)
		}
	}
	return fmt.Errorf("Unknown error parsing input encountered")
}

// Tie the utility functions like title and feed to the fileWriter
func (s *server) fileListener(ctx context.Context, c *fslib.Control) {
	for {
		select {
		case e := <- s.e:
			c.Event(e)
		case j := <- s.j:
			buffs := getChans(j)
			for _, buff := range buffs {
				if ! c.HasBuffer(buff, "feed") {
					s.Open(c, buff)
				}
			}
		case m := <- s.m:
			fileWriter(c, m)
		case <- ctx.Done():			
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

