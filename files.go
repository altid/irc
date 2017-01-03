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
//TODO: Delete member on ctl 
func (srv *Server) Update(s *Show, b *Session) {
	srv.AddFile("main", b.Read(b.Current))
	srv.AddFile("ctl", b.ListFunctions())
	switch {
	case s.Title:
		srv.AddFile("title", "ubqt-irc")
	case s.Status:
		srv.AddFile("status", b.UpdateStatus())
	case s.Tabs:
		srv.AddFile("tabs", b.UpdateTabs())
	case s.Sidebar:
		srv.AddFile("sidebar", b.UpdateSidebar())
	}
}

func (srv *Server) AddFile(key string, file interface{}) {
	if srv.file == nil {
		srv.file = make(map[string]interface{})
	}
	srv.file[key] = file
}
