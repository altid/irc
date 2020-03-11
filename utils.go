package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/altid/libs/config"
	"github.com/altid/libs/fs"
	"github.com/altid/libs/markup"
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
	faside
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
func pm(s *server, msg string) error {
	token := strings.Fields(msg)
	m := &irc.Message{
		Command: "PRIVMSG",
		Prefix: &irc.Prefix{
			Name: s.conf.Name,
		},
		Params: token[:1],
	}
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

func buildConfig() (*config.Config, error) {
	return nil, nil
}

func status(s *server, m *irc.Message) {
	// Just use m.Params[0] for the fname
}

func errorWriter(c *fs.Control, err error) {
	ew, _ := c.ErrorWriter()
	defer ew.Close()

	fmt.Fprintf(ew, "ircfs: %s\n", err)
}

func fileWriter(c *fs.Control, m *msg) error {
	if m.from == "freenode-connect" {
		return nil
	}
	var w *fs.WriteCloser
	var err error
	switch m.fn {
	case fbuffer, faction, fhighlight, fselfaction, fself, ftime:
		w, err = c.MainWriter(m.buff, "feed")
		if err != nil {
			return err
		}

		feed := markup.NewCleaner(w)
		defer feed.Close()

		var color *markup.Color
		switch m.fn {
		case fselfaction:
			color, err = markup.NewColor(markup.Grey, []byte(m.from))
			feed.WritefEscaped(" * %s: ", color)
		case fself:
			color, err = markup.NewColor(markup.Grey, []byte(m.from))
			feed.WritefEscaped("%s: ", color)
		case fbuffer:
			color, err = markup.NewColor(markup.Blue, []byte(m.from))
			feed.WritefEscaped("%s: ", color)
		case faction:
			color, err = markup.NewColor(markup.Blue, []byte(m.from))
			feed.WritefEscaped(" * %s: ", color)
		case fhighlight:
			color, err = markup.NewColor(markup.Red, []byte(m.from))
			feed.WritefEscaped("%s: ", color)
		case ftime:
			color, err = markup.NewColor(markup.Orange, []byte(m.from))
			feed.WritefEscaped("Topic was set by %s, on ", color)
		}

		feed.WritefEscaped("%s\n", m.data)
		return err
	case fnotification:
		ntfy := markup.NewNotifier(m.buff, m.from, m.data)
		c.Notification(ntfy.Parse())

		return nil
	case fserver:
		w, err = c.MainWriter("server", "feed")
	case fstatus:
		w, err = c.StatusWriter(m.buff)
	case faside:
		w, err = c.SideWriter(m.buff)
	case ftitle:
		w, err = c.TitleWriter(m.buff)
	default:
		return nil
	}

	if err != nil {
		return err
	}

	cleaner := markup.NewCleaner(w)
	defer cleaner.Close()

	cleaner.WriteStringEscaped(m.data + "\n")
	return nil
}
