package main

import (
	"github.com/vaughan0/go-ini"
)

func setupShow(conf ini.File, section string, srv *Server) *show {

	//TODO:srv.AddFile("main", current.buf)
	srv.AddFile("main", "test")
	//TODO:srv.AddFile("main", completionList)
	srv.AddFile("ctl", "test")
	show := new(show)
	for key, value := range conf[section] {
		switch key {
		case "Title":
			show.Title = (value == "show")
			//TODO: srv.AddFile("title", usefultitle)
			srv.AddFile("title", "ubqt-irc")
		case "Status":
			show.Status = (value == "show")
			//TODO: srv.AddFile("status", mode/buffer/etc)
			srv.AddFile("status", "test")
		case "Tabs":
			show.Tabs = (value == "show")
			//TODO: srv.AddFile("tabs", bufferlist)
			srv.AddFile("tabs", "test")
		case "Input":
			show.Input = (value == "show")
			srv.AddFile("input", "")
		case "Sidebar":
			show.Sidebar = (value == "show")
			//TODO: srv.Addfile("sidebar", nicklist)
			srv.AddFile("sidebar", "test")
		case "Timestamps":
			show.Timestamps = (value == "show")
		}
	}
	return show
}

func (srv *Server) AddFile(key string, file interface{}) {
	if srv.file == nil {
		srv.file = make(map[string]interface{})
	}
	srv.file[key] = file
}
