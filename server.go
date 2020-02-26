package main

import (
	"bytes"
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
	err := c.CreateBuffer(name, "feed")
	if err != nil {
		return err
	}
	if name[0] == '#' {
		_, err = fmt.Fprintf(s.conn, "JOIN %s\n", name)
		if err != nil {
			return err
		}
	}
	input, err := fs.NewInput(s, workdir, name)
	if err != nil {
		return err
	}
	defer c.Event(path.Join(workdir, name, "input"))
	go input.Start()
	return nil
}

func (s *server) Close(c *fs.Control, name string) error {
	err := c.DeleteBuffer(name, "feed")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.conn, "PART %s\n", name)
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
		err := c.CreateBuffer(t[0], "feed")
		if err != nil {
			return err
		}
		go func() {
			input, _ := fs.NewInput(s, workdir, t[0])
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
func (s *server) Handle(bufname string, l *markup.Lexer) error {
	var m bytes.Buffer
	for {
		i := l.Next()
		switch i.ItemType {
		case markup.EOF:
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
		case markup.ErrorText:
			return fmt.Errorf("error parsing input: %v", i.Data)
		case markup.UrlLink, markup.UrlText, markup.ImagePath, markup.ImageLink, markup.ImageText:
			continue
		case markup.ColorText, markup.ColorTextBold:
			m.WriteString(getColors(i.Data, l))
		case markup.BoldText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case markup.EmphasisText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case markup.UnderlineText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		default:
			m.Write(i.Data)
		}
	}
}

func getColors(current []byte, l *markup.Lexer) string {
	var text bytes.Buffer
	var color bytes.Buffer
	text.Write(current)
	for {
		i := l.Next()
		switch i.ItemType {
		case markup.EOF:
			return color.String()
		case markup.ColorCode:
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
		case markup.ColorTextBold:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case markup.ColorTextEmphasis:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case markup.ColorTextUnderline:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case markup.ColorText:
			text.Write(i.Data)
		}
	}
}

func getColorCode(d []byte) string {
	switch string(d) {
	case markup.White:
		return "0"
	case markup.Black:
		return "1"
	case markup.Blue:
		return "2"
	case markup.Green:
		return "3"
	case markup.Red:
		return "4"
	case markup.Brown:
		return "5"
	case markup.Purple:
		return "6"
	case markup.Orange:
		return "7"
	case markup.Yellow:
		return "8"
	case markup.LightGreen:
		return "9"
	case markup.Cyan:
		return "10"
	case markup.LightCyan:
		return "11"
	case markup.LightBlue:
		return "12"
	case markup.Pink:
		return "13"
	case markup.Grey:
		return "14"
	case markup.LightGrey:
		return "15"
	}
	return ""
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
			err := fileWriter(c, m)
			if err != nil {
				errorWriter(c, err)
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
	tlsconn.Handshake()
	s.conn = tlsconn
	return nil
}
