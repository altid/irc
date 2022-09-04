package session

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/altid/libs/markup"
	"github.com/altid/libs/service/controller"
	"gopkg.in/irc.v3"
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
func pm(s *Session, msg string) error {
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

func action(s *Session, from, msg string) error {
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

func sendmsg(s *Session, m *irc.Message) error {
	w := irc.NewWriter(s.conn)
	return w.WriteMessage(m)
}

func timeSetAt(s *Session, m *irc.Message) {
	i, err := strconv.ParseInt(m.Params[3], 10, 64)
	if err != nil {
		return
	}
	t := time.Unix(i, 0).Format(time.RFC1123)
	from := strings.Split(m.Params[2], "!")
	s.m <- &msg{
		buff: m.Params[1],
		data: t,
		from: from[0],
		fn:   ftime,
	}
}

func title(name string, s *Session, m *irc.Message) {
	s.m <- &msg{
		buff: name,
		data: m.Trailing(),
		fn:   ftitle,
	}
}

func feed(fn fname, name string, s *Session, m *irc.Message) {
	s.m <- &msg{
		buff: name,
		data: m.Trailing(),
		from: m.Prefix.Name,
		fn:   fn,
	}
}

func status(s *Session, m *irc.Message) {
	s.m <- &msg{
		buff: m.Params[0],
		data: m.Params[1],
		fn:   fstatus,
	}
}

func errorWriter(c controller.Controller, err error) {
	ew, _ := c.ErrorWriter()
	defer ew.Close()

	fmt.Fprintf(ew, "ircfs: %s\n", err)
}

func fileWriter(c controller.Controller, m *msg) error {
	switch m.fn {
	case fbuffer, faction, fhighlight, fselfaction, fself, ftime:
		return m.fnormalWrite(c)
	case fnotification:
		return c.Notification(markup.NewNotifier(m.buff, m.from, m.data).Parse())
	case fserver:
		return m.fspecialWrite(c.FeedWriter("server"))
	case fstatus:
		return m.fspecialWrite(c.StatusWriter(m.buff))
	case faside:
		return m.fspecialWrite(c.SideWriter(m.buff))
	case ftitle:
		return m.fspecialWrite(c.TitleWriter(m.buff))
	default:
		return nil
	}
}

// We take the error in here for a cleaner switch
func (m *msg) fspecialWrite(w controller.WriteCloser, err error) error {
	if err != nil {
		return fmt.Errorf("error in special writer: %s", err)
	}

	cleaner := markup.NewCleaner(w)
	defer cleaner.Close()

	if _, e := cleaner.WriteStringEscaped(m.data + "\n"); e != nil {
		return e
	}

	return nil
}

func (m *msg) fnormalWrite(c controller.Controller) error {
	var color *markup.Color

	w, err := c.FeedWriter(m.buff)
	if err != nil {
		fmt.Printf("fnormalwrite: %s\n", err)
		return err
	}

	feed := markup.NewCleaner(w)
	defer feed.Close()

	switch m.fn {
	case fselfaction:
		color, _ = markup.NewColor(markup.Grey, []byte(m.from))
		feed.WritefEscaped(" * %s: ", color)
	case fself:
		color, _ = markup.NewColor(markup.Grey, []byte(m.from))
		feed.WritefEscaped("%s: ", color)
	case fbuffer:
		color, _ = markup.NewColor(markup.Blue, []byte(m.from))
		feed.WritefEscaped("%s: ", color)
	case faction:
		color, _ = markup.NewColor(markup.Blue, []byte(m.from))
		feed.WritefEscaped(" * %s: ", color)
	case fhighlight:
		color, _ = markup.NewColor(markup.Red, []byte(m.from))
		feed.WritefEscaped("%s: ", color)
	case ftime:
		color, _ = markup.NewColor(markup.Orange, []byte(m.from))
		feed.WritefEscaped("Topic was set by %s, on ", color)
	}

	_, err = feed.WritefEscaped("%s\n", m.data)
	return err
}
