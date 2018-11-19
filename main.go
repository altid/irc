package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/go-irc/irc"
	"github.com/mischief/ndb"
)

var (
	inPath = flag.String("p", "irc", "path for filesystem - can be relative to home, or complete path to existing directory")
	config = flag.String("c", "config", "Configuration file")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	conf, err := ndb.Open(*config)
	if err != nil {
		log.Fatal(err)
	}
	formats := GetFormats(conf)
	servers := GetServers(conf)
	os.MkdirAll(*inPath, 0755)
	defer os.RemoveAll(*inPath)
	for key, srv := range servers {
		srv.conf.Handler = srv.InitHandlers(formats)
		addr := srv.addr + ":" + srv.port
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Println(err)
			delete(servers, key)
			continue
		}
		client := irc.NewClient(conn, srv.conf)
		srv.client = client
		//go client.Run()
		client.Run()
	}
	//ControlLoop(servers)
}
