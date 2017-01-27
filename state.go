package main

import (
	"github.com/vaughan0/go-ini"
)

func parseOptions(st *State, conf ini.File) {
	st.show = make(map[string]bool)
	for key, value := range conf["options"] {
		switch key {
		case "Title":
			st.show["title"] = (value == "show")
		case "Status":
			st.show["status"] = (value == "show")
		case "Tabs":
			st.show["tabs"] = (value == "show")
		case "Input":
			st.input = make(chan string)
			st.show["input"] = (value == "show")
		case "Sidebar":
			st.show["sidebar"] = (value == "show")
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
	st.show["ctl"] = true
	st.show["main"] = true
	for section := range conf {
		if section == "options" {
			continue
		}
		// Fires off IRC sessions
		setupServer(conf, section, st)
		st.buffer = section
		st.server = section
	}
	return nil
}
