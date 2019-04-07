package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/altid/cleanmark"
	"github.com/altid/fslib"
	"github.com/go-irc/irc"
)

type fname int

const (
	faction fname = iota
	fbuffer
	fhighlight
	fnotification
	fself
	fselfaction
	fserver
	fsidebar
	fstatus
	ftime
	ftitle
)

type msg struct {
	buff string
	data string
	from string
	fn   fname
}

func getChans(buffs string) []string {
	var items []string
	r := csv.NewReader(strings.NewReader(buffs))
	for {
		buffers, err := r.Read()
		if err == io.EOF {
			break
		}
		items = append(items, buffers...)
	}
	return items
}

// Private message
// TODO(halfwit): We need to create a buffer if we're initializing a PM
// `open`ing a conversation with a user does as well
// https://github.com/altid/ircfs/issues/6
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

func action(s *server, from, msg string) error {
	m := &irc.Message{
		Command: "PRIVMSG",
		Prefix: &irc.Prefix{
			Name: s.conf.Name,
		},
		Params: []string{
			from,
			fmt.Sprintf("ACTION %s", msg),
		},
	}
	return sendmsg(s, m)
}

func sendmsg(s *server, m *irc.Message) error {
	w := irc.NewWriter(s.conn)
	return w.WriteMessage(m)
}

func timeSetAt(s *server, m *irc.Message) {
	i, err := strconv.ParseInt(m.Params[3], 10, 64)
	if err != nil {
		return
	}
	t := time.Unix(i, 0).Format(time.RFC1123)
	from := strings.Split(m.Params[2], "!")
	s.m <- &msg{
		buff: m.Params[1],
		data: fmt.Sprintf("%s", t),
		from: from[0],
		fn:   ftime,
	}
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
		buff: name,
		data: m.Trailing(),
		from: m.Prefix.Name,
		fn:   fn,
	}
}

func status(s *server, m *irc.Message) {
	// Just use m.Params[0] for the fname
}

func fileWriter(c *fslib.Control, m *msg) {
	if m.from == "freenode-connect" {
		return
	}
	var w *fslib.WriteCloser
	switch m.fn {
	case fbuffer, faction, fhighlight, fselfaction, fself, ftime:
		w = c.MainWriter(m.buff, "feed")
		if w == nil {
			return
		}
		feed := cleanmark.NewCleaner(w)
		defer feed.Close()
		switch m.fn {
		case fselfaction:
			color, _ := cleanmark.NewColor(cleanmark.Grey, []byte(m.from))
			feed.WritefEscaped(" * %s: ", color)
		case fself:
			color, _ := cleanmark.NewColor(cleanmark.Grey, []byte(m.from))
			feed.WritefEscaped("%s: ", color)
		case fbuffer:
			color, _ := cleanmark.NewColor(cleanmark.Blue, []byte(m.from))
			feed.WritefEscaped("%s: ", color)
		case faction:
			color, _ := cleanmark.NewColor(cleanmark.Blue, []byte(m.from))
			feed.WritefEscaped(" * %s: ", color)
		case fhighlight:
			color, _ := cleanmark.NewColor(cleanmark.Red, []byte(m.from))
			feed.WritefEscaped("%s: ", color)
		case ftime:
			color, _ := cleanmark.NewColor(cleanmark.Orange, []byte(m.from))
			feed.WritefEscaped("Topic was set by %s, on ", color)
		}
		feed.WritefEscaped("%s\n", m.data)
		return
	// TODO halfwit: clean m.data and m.from
	case fnotification:
		ntfy := cleanmark.NewNotifier(m.buff, m.from, m.data)
		c.Notification(ntfy.Parse())
	case fserver:
		w = c.MainWriter("server", "feed")
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
	cleaner.WriteStringEscaped(m.data + "\n")
}
