package main

import (
	"github.com/vaughan0/go-ini"
)

// Function parses the options in the irc.ini to give defaults 
// Subsequent clients connecting in will not modify
// the State of any other client
func setupState(conf ini.File, section string, s *State) {

	for key, value := range conf[section] {
		switch key {
		case "Title":
			s.Title = (value == "show")
		case "Status":
			s.Status = (value == "show")
		case "Tabs":
			s.Tabs = (value == "show")
		case "Input":
			s.Input = (value == "show")
		case "Sidebar":
			s.Sidebar = (value == "show")
		case "Timestamps":
			s.Timestamps = (value == "show")
		}
	}
}
