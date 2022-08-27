package ircfs

import (
	"context"

	"github.com/altid/ircfs/internal/commands"
	"github.com/altid/ircfs/internal/session"
	"github.com/altid/libs/config"
	"github.com/altid/libs/mdns"
	"github.com/altid/libs/service"
	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/store"
)

type Ircfs struct {
	run     func() error
	session *session.Session
	name    string
	debug   bool
	mdns    *mdns.Entry
	ctx     context.Context
}

func CreateConfig(srv string, debug bool) error {
	d := &session.Defaults{}
	return config.Create(d, srv, "", debug)
}

// This connects to IRC, manages interactions with the plugins
func Register(ssh, ldir bool, addr, srv string, debug bool) (*Ircfs, error) {
	// Some sane-ish defaults
	defaults := &session.Defaults{
		Address: "irc.libera.chat",
		SSL:     "none",
		Port:    6667,
		Auth:    "password",
		Filter:  "",
		Nick:    "guest",
		User:    "guest",
		Name:    "guest",
		Buffs:   "#altid",
		Logdir:  "",
		TLSCert: "",
		TLSKey:  "",
	}

	if e := config.Marshal(defaults, srv, "", debug); e != nil {
		return nil, e
	}

	l, err := tolisten(defaults, addr, ssh, debug)
	if err != nil {
		return nil, err
	}

	s := tostore(defaults, ldir)
	session := &session.Session{
		Defaults: defaults,
		Verbose:  debug,
	}

	session.Parse()
	ctx := context.Background()

	i := &Ircfs{
		session: session,
		ctx:     ctx,
		name:    srv,
		debug:   debug,
	}

	c := service.New(srv, addr, debug)
	c.WithListener(l)
	c.WithStore(s)
	c.WithContext(ctx)
	//c.WithCallbacks()
	c.WithRunner(session)

	// Add in commands and make sure our type has a controller as well
	c.SetCommands(commands.Commands)
	i.run = c.Listen

	return i, nil
}

func (ircfs *Ircfs) Run() error {
	return ircfs.run()
}

func (ircfs *Ircfs) Broadcast() {
	entry := &mdns.Entry{
		Addr: ircfs.session.Defaults.Address,
		Name: ircfs.name,
		Txt:  nil,
		Port: ircfs.session.Defaults.Port,
	}

	mdns.Register(entry)
	ircfs.mdns = entry
}

func (ircfs *Ircfs) Cleanup() {
	if ircfs.mdns != nil {
		ircfs.mdns.Cleanup()
	}
	ircfs.session.Quit()
}

func (ircfs *Ircfs) Session() *session.Session {
	return ircfs.session
}

func tolisten(d *session.Defaults, addr string, ssh, debug bool) (listener.Listener, error) {
	//if ssh {
	//    return listener.NewListenSsh()
	//}

	if d.TLSKey == "none" && d.TLSCert == "none" {
		return listener.NewListen9p(addr, "", "", debug)
	}

	return listener.NewListen9p(addr, d.TLSCert, d.TLSKey, debug)
}

func tostore(d *session.Defaults, ldir bool) store.Filer {
	if ldir {
		return store.NewLogStore(d.Logdir.String())
	}

	return store.NewRamStore()
}
