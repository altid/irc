package ircfs

import (
	"context"

	"github.com/altid/ircfs/internal/session"
	"github.com/altid/libs/config"
	"github.com/altid/libs/mdns"
	"github.com/altid/libs/service"
	"github.com/altid/libs/service/listener"
	"github.com/altid/libs/store"
)

type Ircfs struct {
    session     *session.Session
    control     *service.Control
    listener    listener.Listener
    store       store.Filer
    name        string
    debug       bool
    mdns        *mdns.Entry
    ctx         context.Context
}

func CreateConfig(srv string, debug bool) error {
    d := &session.Defaults{}
    return config.Create(d, srv, "", debug)
}

// This connects to IRC, manages interactions with the plugins 
func Register(ssh, ldir bool, srv string, debug bool) (*Ircfs, error) {
    // Some sane-ish defaults
    defaults := &session.Defaults{
        Address: "libera.chat",
        SSL:     "none",
        Port:    6697,
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

    l, err := tolisten(defaults, ssh)
    if err != nil { 
        return nil, err
    }

    s:= tostore(defaults, ldir)
    session := &session.Session{
        Defaults: defaults,
        Verbose: debug,
    }

    session.Parse()
    ctx := context.Background()

    i := &Ircfs{
        session: session,
        store: s,
        listener: l,
        ctx: ctx,
        name: srv,
        debug: debug,
    }

    c, err := service.New(i.session, s, l, defaults.Logdir.String(), debug)
    if err != nil {
        return nil, err
    }

    i.control = c
    return i, nil
}

func (ircfs *Ircfs) Broadcast() {
    entry := &mdns.Entry{
        Addr: ircfs.session.Defaults.Address,
        Name: ircfs.name,
        Txt: nil,
        Port: ircfs.session.Defaults.Port,
    }

    mdns.Register(entry)
    ircfs.mdns = entry
}
// Start connects to IRC and does the things
func (ircfs *Ircfs) Start() error {
    return ircfs.session.Start(ircfs.ctx, ircfs.control)
}

// Listen starts up our listener on the network
func (ircfs *Ircfs) Listen() error {
    return ircfs.listener.Listen()
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

func tolisten(d *session.Defaults, ssh bool) (listener.Listener, error) {
    //if ssh {
    //    return listener.NewListenSsh()
    //}

    return listener.NewListen9p(d.Address, d.TLSCert, d.TLSKey)
}

func tostore(d *session.Defaults, ldir bool) store.Filer {
    if ldir {
        return store.NewLogStore(d.Logdir.String())
    }

    return store.NewRamStore()
}