package ircfs

import (
	"context"

	"github.com/altid/ircfs/internal/commands"
	"github.com/altid/ircfs/internal/session"
	"github.com/altid/libs/config"
	"github.com/altid/libs/service"
)

type Ircfs struct {
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
func Register(srv string, debug bool) (*Ircfs, error) {
	if e := config.Marshal(defaults, srv, "", debug); e != nil {
		return nil, e
	}

	session := &session.Session{
		Defaults: defaults,
		Verbose:  debug,
	}

	ctx := context.Background()
	session.Parse(ctx)

	i := &Ircfs{
		session: session,
		ctx:     ctx,
		name:    srv,
		debug:   debug,
	}
	
	svc, err := service.Register(ctx, srv)
	if err != nil {
		return nil, err
	}

	svc.SetCallbacks(session)
	svc.SetCommands(commands.Commands)
	i.run = svc.Listen
	return i, nil
}

func (ircfs *Ircfs) Run() error {
	return ircfs.run()
}

func (ircfs *Ircfs) Cleanup() {
	ircfs.session.Quit()
}

func (ircfs *Ircfs) Session() *session.Session {
	return ircfs.session
}
