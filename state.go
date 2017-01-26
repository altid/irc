package main

import (
	"github.com/vaughan0/go-ini"
)

func parseOptions(st *State, conf ini.File) {
	for key, value := range conf["options"] {
		st.show = make(map[string]bool)
		switch key {
		case "Title":
			st.show["Title"] = (value == "show")
		case "Status":
			st.show["Status"] = (value == "show")
		case "Tabs":
			st.show["Tabs"] = (value == "show")
		case "Input":
			st.input = make(chan string)
			st.show["Input"] = (value == "show")
		case "Sidebar":
			st.show["Bar"] = (value == "show")
		}
	}
}

// Initialize - Read config and set up IRC sessions per entry
func (st *State) Initialize() error {
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		return err
	}
	parseOptions(st, conf)
	for section := range conf {
		if section == "options" {
			continue
		}
		// Fires off IRC sessions
		setupServer(conf, section, st)
		setupIrc(conf, section, st)
		st.buffer = section
		st.server = section
	}
	return nil
}
