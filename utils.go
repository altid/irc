package main

import (
	"strings"

	"github.com/go-irc/irc"
)

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
