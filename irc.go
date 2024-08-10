package irc

import (
	"context"

	"github.com/altid/irc/internal/commands"
	"github.com/altid/irc/internal/session"
	"github.com/altid/libs/config"
	"github.com/altid/libs/service"
)

type Irc struct {
	run		func() error
	session *session.Session
	name    string
	debug   bool
	ctx     context.Context
}

// Some sane-ish defaults
var defaults *session.Defaults = &session.Defaults{
	Address: "irc.libera.chat",
	SSL:     "none",
	Port:    6667,
	Filter:  "",
	Nick:    "guest",
	User:    "guest",
	Name:    "guest",
	Buffs:   "#altid",
	TLSCert: "",
	TLSKey:  "",
}

func CreateConfig(srv string, debug bool) error {
	return config.Create(defaults, srv, "", debug)
}

// This connects to IRC, manages interactions with the plugins
func Register(srv string, fg, debug bool) (*Irc, error) {
	if e := config.Marshal(defaults, srv, "", debug); e != nil {
		return nil, e
	}

	session := &session.Session{
		Defaults: defaults,
		Verbose:  debug,
	}

	ctx := context.Background()
	session.Parse(ctx)

	i := &Irc{
		session: session,
		ctx:     ctx,
		name:    srv,
		debug:   debug,
	}
	
	svc, err := service.Register(ctx, srv, fg)
	if err != nil {
		return nil, err
	}

	svc.SetCallbacks(session)
	svc.SetCommands(commands.Commands)
	i.run = svc.Listen
	return i, nil
}

func (irc *Irc) Run() error {
	return irc.run()
}

func (irc *Irc) Cleanup() {
	irc.session.Quit()
}

func (irc *Irc) Session() *session.Session {
	return irc.session
}
