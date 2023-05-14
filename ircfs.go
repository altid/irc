package ircfs

import (
	"context"
	"net/url"
	"strconv"

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
	addr    string
	debug   bool
	mdns    *mdns.Entry
	ctx     context.Context
}

// Some sane-ish defaults
var defaults *session.Defaults = &session.Defaults{
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

func CreateConfig(srv string, debug bool) error {
	return config.Create(defaults, srv, "", debug)
}

// This connects to IRC, manages interactions with the plugins
func Register(ssh, ldir bool, addr, srv string, debug bool) (*Ircfs, error) {
	if e := config.Marshal(defaults, srv, "", debug); e != nil {
		return nil, e
	}

	l, err := tolisten(defaults, addr, ssh, debug)
	if err != nil {
		return nil, err
	}

	s := tostore(defaults, ldir, debug)
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
		addr:    addr,
		debug:   debug,
	}

	c := service.New(srv, addr, debug)
	c.WithListener(l)
	c.WithStore(s)
	c.WithContext(ctx)
	c.WithCallbacks(session)
	c.WithRunner(session)

	// Add in commands and make sure our type has a controller as well
	c.SetCommands(commands.Commands)
	i.run = c.Listen

	return i, nil
}

func (ircfs *Ircfs) Run() error {
	return ircfs.run()
}

func (ircfs *Ircfs) Broadcast() error {
	dial, err := url.Parse(ircfs.addr)
	if err != nil {
		return err
	}
	entry := &mdns.Entry{
		Addr: dial.Hostname(),
		Name: ircfs.name,
		Port: 564,
	}
	if(dial.Port() != "") {
		entry.Port, _ = strconv.Atoi(dial.Port())
	}

	if e := mdns.Register(entry); e != nil {
		return e
	}

	ircfs.mdns = entry
	return nil
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

func tostore(d *session.Defaults, ldir, debug bool) store.Filer {
	if ldir {
		return store.NewLogStore(d.Logdir.String(), debug)
	}

	return store.NewRamStore(debug)
}
