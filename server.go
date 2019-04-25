package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path"
	"strings"

	cm "github.com/altid/cleanmark"
	"github.com/altid/fslib"
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
		e:      e,
		m:      m,
		j:      j,
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
	if err != nil {
		return nil
	}
	input, err := fslib.NewInput(s, workdir, name)
	if err != nil {
		return err
	}
	defer c.Event(path.Join(workdir, name, "input"))
	go input.Start()
	return nil
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

func (s *server) Default(c *fslib.Control, cmd, from, m string) error {
	switch cmd {
	case "a", "act", "action", "me":
		return action(s, from, m)
	case "msg", "query": 
		// we don't want to send a JOIN message, so we don't simply s.j <- t[0]
		t := strings.Fields(m)
		err := c.CreateBuffer(t[0], "feed")
		if err != nil {
			return err
		}
		go func() {
			input, _ := fslib.NewInput(s, workdir, t[0])
			input.Start()
		}()
		return pm(s, m)
	case "nick":
		// Make sure we update s.conf.Name when we update username
		s.conf.Name = m
		fmt.Fprintf(s.conn, "NICK %s\n", m)
		return nil
	}
	return fmt.Errorf("Unknown command %s", cmd)
}

// input is always sent down raw to the server
func (s *server) Handle(bufname string, l *cm.Lexer) error {
	var m strings.Builder
	for {
		i := l.Next()
		switch i.ItemType {
		case cm.EOF:
			b := path.Base(bufname)
			d := m.String()
			_, err := fmt.Fprintf(s.conn, ":%s PRIVMSG %s :%s\n", s.conf.Name, b, d)
			s.m <- &msg{
				buff: b,
				from: s.conf.Nick,
				data: d,
				fn:   fself,
			}
			return err
		case cm.ErrorText:
			return fmt.Errorf("Error parsing input: %v", i.Data)
		case cm.UrlLink, cm.UrlText, cm.ImagePath, cm.ImageLink, cm.ImageText:
			continue
		case cm.ColorText, cm.ColorTextBold:
			m.WriteString(getColors(i.Data, l))
		case cm.BoldText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case cm.EmphasisText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case cm.UnderlineText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		default:
			m.Write(i.Data)
		}
	}
	return fmt.Errorf("Unknown error parsing input encountered")
}

func getColors(current []byte, l *cm.Lexer) string {
	var text strings.Builder
	var color strings.Builder
	text.Write(current)
	for {
		i := l.Next()
		switch i.ItemType {
		case cm.EOF:
			return color.String()
		case cm.ColorCode:
			code := getColorCode(i.Data)
			if n := bytes.IndexByte(i.Data, ','); n >= 0 {
				code = getColorCode(i.Data[:n])
				code += ","
				code += getColorCode(i.Data[n+1:])	
			}
			color.WriteString("")
			color.WriteString(code)
			color.WriteString(text.String())
			color.WriteString("")
			return color.String()
		case cm.ColorTextBold:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case cm.ColorTextEmphasis:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case cm.ColorTextUnderline:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case cm.ColorText:
			text.Write(i.Data)
		}
	}
}

func getColorCode(d []byte) string {
	switch string(d) {
	case cm.White:
		return "0"
	case cm.Black:
		return "1"
	case cm.Blue:
		return "2"
	case cm.Green:
		return "3"
	case cm.Red:
		return "4"
	case cm.Brown:
		return "5"
	case cm.Purple:
		return "6"
	case cm.Orange:
		return "7"
	case cm.Yellow:
		return "8"
	case cm.LightGreen:
		return "9"
	case cm.Cyan:
		return "10"
	case cm.LightCyan:
		return "11"
	case cm.LightBlue:
		return "12"
	case cm.Pink:
		return "13"
	case cm.Grey:
		return "14"
	case cm.LightGrey:
		return "15"
	}
	return ""
}

// Tie the utility functions like title and feed to the fileWriter
func (s *server) fileListener(ctx context.Context, c *fslib.Control) {
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
			fileWriter(c, m)
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
	tlsconn.Handshake()
	s.conn = tlsconn
	return nil
}
