package main

import (
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
		// Fires off IRC sessions
		st.buffer = section
		st.server = section
	}
	return nil
}
