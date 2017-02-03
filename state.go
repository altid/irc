package main

import (
	//"github.com/lrstanley/girc"
	"github.com/ubqt-systems/ubqtlib"
	"github.com/vaughan0/go-ini"
)

func parseOptions(srv *ubqtlib.Srv, conf ini.File) {
	for key, value := range conf["options"] {
		if value == "show" {
			srv.AddFile(key)
		}
	}
}

// Initialize - Read config and set up IRC sessions per entry
func (st *State) initialize(srv *ubqtlib.Srv) error {
	//st.ctl = getCtl()
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		return err
	}
	parseOptions(srv, conf)
	srv.AddFile("ctl")
	srv.AddFile("main")
	for section := range conf {
		if section == "options" {
			continue
		}
		/*
			conf := girc.Config {
				Server: Server,
				Port: Port,
				Nick: Nick,
				User: User,
				Name: Name,
				MaxRetries: 3,
				Logger: os.Stdout,
			}
			client := girc.New(conf)
			client.Callbacks.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
				c.Join(channels)
			})
			client.Callbacks.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
				//TODO: handle privmsg
			})
			err = client.Connect()
			if err != nil {
				log.Fatalf("an error occured while attempting to connect to %s: %s", client.Server(), err)
			}
			// Fire off IRC connection
			go client.Loop()
		*/
		st.buffer = section
		st.server = section
	}
	return nil
}
