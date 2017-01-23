package main

import (
	"fmt"
	"github.com/vaughan0/go-ini"
)

func inputListener(st *state) {
	for {
		select {
		case input := <-st.input.ch:
			fmt.Println(string(input.buf))
		}
	}
}

// Iterate through conf, setting up state
func newState() (*state, error) {
	var st state
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		return nil, err
	}
	st.current = new(Current)
	st.current.buffer = make(map[string]string)
	st.current.server = make(map[string]string)
	st.current.ch = make(chan int, 10)
	st.bar = new(Sidebar)
	st.bar.names = make(map[string]string)
	st.title = new(Title)
	st.status = new(Status)
	st.tabs = new(Tabs)
	st.input = new(Input)
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
					st.input.ch = make(chan *Message)
					st.input.show = (value == "show")
				case "Sidebar":
					st.bar.show = (value == "show")
				case "Timestamps":
					st.timestamps = make(map[string]bool)
					st.timestamps["main"] = (value == "show")
				}
			}
			continue
		}
		setupServer(conf, section, &st)

		st.current.buffer["main"] = section
		st.current.server["main"] = section
		st.bar.names["main"] = section
	}
	go inputListener(&st)
	return &st, nil
}
