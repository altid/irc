package ircfs

import (
	"context"

	"github.com/altid/ircfs/internal/session"
	"github.com/altid/libs/config"
	"github.com/altid/libs/service"
)

type Ircfs struct {
	run     func() error
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
	_, err := service.Start(srv, session)
	if err != nil {
		return nil, err
	}

	// Literally just write to C? But we should wrap in convenience now after.
	// Add Commands
	// Add Context
/*
	c.WithContext(ctx)

	// Add in commands and make sure our type has a controller as well
	c.SetCommands(commands.Commands)
*/
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
