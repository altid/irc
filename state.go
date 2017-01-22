package main

import (
	"github.com/vaughan0/go-ini"
)

// Iterate through conf, setting up state
// launching IRC goroutines
func newState() (*state, error) {
	var st state
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		return nil, err
	}
	st.current = new(Current)
	st.title = new(Title)
	st.status = new(Status)
	st.tabs = new(Tabs)
	st.input = new(Input)
	st.bar = new(Sidebar)
	st.ctl = new(Ctl)
	for section := range conf {
		if section == "options" {
			for key, value := range conf[section] {
				switch key {
				case "Title":
					st.title.show = (value == "show")
				case "Status":
					st.status.show = (value == "show")
				case "Tabs":
					st.tabs.show = (value == "show")
				case "Input":
					st.input.show = (value == "show")
				case "Sidebar":
					st.bar.show = (value == "show")
				case "Timestamps":
					st.timestamps = (value == "show")
				}
			}
			continue
		}
		setupServer(conf, section, &st)
		st.current.buffer = section
		st.current.server = section
	}
	return &st, nil
}
