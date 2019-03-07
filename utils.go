package main

import (
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
	fsidebar
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

func title(fileName string, s *server, m *irc.Message) {
	s.m <- &msg{
		buff: fileName,
		data: m.Trailing(),
		fn:   ftitle,
		
	}
}

func feed(mtype messageType, name string, s *server, m *irc.Message) {
	var data string
	switch mtype {
	case chanMsg:
		data = "foo"
	case serverMsg:
		data = "bar"
	default:
		return
	}
	s.m <- &msg{
		buff: name,
		data: data,
		from: m.Prefix.Name,
		fn:   fbuffer,
	}
}

func status(s *server, m *irc.Message) {
	// Just use m.Params[0] for the fname
}

// Probably switch this to an iota, since we have access to both ends of this intra file
func fileWriter(c *fslib.Control, m *msg) {
	var w *fslib.WriteCloser
	switch m.fn {
	case fbuffer:
		// Create and link buffer if it isn't present
		w = c.MainWriter(m.buff, "feed")
	case fstatus:
		w = c.StatusWriter(m.buff)
	//case fsidebar:
	case ftitle:
		w = c.TitleWriter(m.buff)
	default:
		return
	}
	cleaner := cleanmark.NewCleaner(w)
	defer cleaner.Close()
	cleaner.WriteStringEscaped(m.data + "\n")
}
