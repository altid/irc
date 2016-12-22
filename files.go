package main

import (
	"github.com/vaughan0/go-ini"
)

func setupShow(conf ini.File, section string, srv *server) *show {
	show := new(show)
	for key, value := range conf[section] {
		switch key {
		case "Title":
			show.Title = (value == "show")
		case "Status":
			show.Status = (value == "show")
		case "Tabs":
			show.Tabs = (value == "show")
		case "Input":
			show.Input = (value == "show")
		case "Sidebar":
			show.Sidebar = (value == "show")
		case "Timestamps":
			show.Timestamps = (value == "show")
		}
	}
	return show
}
