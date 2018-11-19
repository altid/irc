package main

import (
	"github.com/go-irc/irc"
	"github.com/mischief/ndb"
)

type Server struct {
	addr     string
	port     string
	conf     irc.ClientConfig
	client   *irc.Client
	ctl      chan string
	channels []string
	filter   string
}

func GetServers(ndb *ndb.Ndb) map[string]*Server {
	servers := make(map[string]*Server)
	for _, rec := range ndb.Search("service", "irc") {
		ctl := make(chan string)
		conf := &irc.ClientConfig{}
		server := &Server{port: "6667", ctl: ctl, conf: *conf, filter: "none"}
		for _, tup := range rec {
			switch tup.Attr {
			case "address":
				server.addr = tup.Val
			case "port":
				server.port = tup.Val
			case "filter":
				server.filter = tup.Val
			case "channels":
				server.channels = append(server.channels, tup.Val)
			case "nick":
				server.conf.Nick = tup.Val
			case "password":
				server.conf.Pass = tup.Val
			case "user":
				server.conf.User = tup.Val
			case "name":
				server.conf.Name = tup.Val
			}
		}
		servers[server.addr] = server
	}
	return servers
}
