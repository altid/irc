package main

import (
	"log"
	"path"
	"strings"

	"github.com/go-irc/irc"
	"github.com/ubqt-systems/cleanmark"
	"github.com/ubqt-systems/fslib"
)

type fname int
const (
	ftitle fname = iota
	fstatus
	fbuffer
	faction
	fsidebar
	fserver
)

type msg struct {
	buff string
	data string
	from string
	fn   fname
}

// Private message
func pm(s *server, msg string) error {
	token := strings.Fields(msg)
	m := &irc.Message{
		Command: "PRIVMSG",
		Prefix: &irc.Prefix{
			Name: s.conf.Name,
		},
		Params: token[:1],
	}
	// Param[1] is the body of the msg
	m.Params = append(m.Params, strings.Join(token[1:], " "))
	return sendmsg(s, m)
}

func sendmsg(s *server, m *irc.Message) error {
	w := irc.NewWriter(s.conn)
	return w.WriteMessage(m)
}

func title(name string, s *server, m *irc.Message) {
	s.m <- &msg{
		buff: name,
		data: m.Trailing(),
		fn:   ftitle,
	}
}

func feed(fn fname, name string, s *server, m *irc.Message) {
	s.m <- &msg{
		buff: path.Join(name, "feed"),
		data: m.Trailing(),
		from: m.Prefix.Name,
		fn:   fn,
	}
}

func status(s *server, m *irc.Message) {
	// Just use m.Params[0] for the fname
}

// Probably switch this to an iota, since we have access to both ends of this intra file
func fileWriter(c *fslib.Control, m *msg) {
	var w *fslib.WriteCloser
	switch m.fn {
	case fbuffer, faction:
		// Here we want to parse for uname(highlights)
		//c.CreateBuffer(
		w = c.MainWriter(m.buff, "feed")
		// Color
		// switch on faction/fbuffer to decide which token
		// WritefEscaped
		// return
	case fserver:
		// Create buffer if not exist
	case fstatus:
		w = c.StatusWriter(m.buff)
	case fsidebar:
		w = c.SideWriter(m.buff)
	case ftitle:
		w = c.TitleWriter(m.buff)
	}
	if w == nil {
		return
	}
	cleaner := cleanmark.NewCleaner(w)
	defer cleaner.Close()
	// if m.from write it
	// write trailing
	cleaner.WriteStringEscaped(m.data + "\n")
}
