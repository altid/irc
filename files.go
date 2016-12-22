package main

import (
	"github.com/vaughan0/go-ini"
)

func setupShow(conf ini.File, section string) *Show {

	Show := new(Show)
	for key, value := range conf[section] {
		switch key {
		case "Title":
			Show.Title = (value == "show")
		case "Status":
			Show.Status = (value == "show")
		case "Tabs":
			Show.Tabs = (value == "show")
		case "Input":
			Show.Input = (value == "show")
		case "Sidebar":
			Show.Sidebar = (value == "show")
		case "Timestamps":
			Show.Timestamps = (value == "show")
		}
	}
	return Show
}

func (srv *Server) Update(s *Show, b *Session) {
	srv.AddFile("main", b.Main)
	//TODO:srv.AddFile("main", completionList)
	srv.AddFile("ctl", "test")
	switch {
	case s.Title:
		//TODO: srv.AddFile("title", usefultitle)
		srv.AddFile("title", "ubqt-irc")
	case s.Status:
		//TODO: srv.AddFile("status", mode/buffer/etc)
		srv.AddFile("status", "test")
	case s.Tabs:
		//TODO: srv.AddFile("tabs", bufferlist)
		srv.AddFile("tabs", "test")
	case s.Input:
		//TODO: srv.AddFile("tabs", scrollback)
		srv.AddFile("input", "stuff")
	case s.Sidebar:
		//TODO: srv.Addfile("sidebar", nicklist)
		srv.AddFile("sidebar", "test")
	}
}

func (srv *Server) AddFile(key string, file interface{}) {
	if srv.file == nil {
		srv.file = make(map[string]interface{})
	}
	srv.file[key] = file
}
