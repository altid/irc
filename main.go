package main

import (
	"flag"
	"log"
	"net"
	"path"
	"os"
	"os/user"

	"github.com/go-irc/irc"
	"github.com/mischief/ndb"
)

var (
	inPath  = flag.String("p", "irc", "path for filesystem - can be relative to home, or complete path to existing directory")
	config  = flag.String("c", "config", "Configuration file")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	// Test inPath exists, else set relative to homedir
	if _, err := os.Stat(*inPath); os.IsNotExist(err) {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		*inPath = path.Join(user.HomeDir, *inPath)
	}
	if _, err := os.Stat(*config); os.IsNotExist(err) {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		*config = path.Join(user.HomeDir, *inPath, *config)
	}
	conf, err := ndb.Open(*config)
	if err != nil {
		log.Fatal(err)
	}
	formats := GetFormats(conf)
	servers := GetServers(conf)
	for key, srv := range servers {
		srv.conf.Handler = srv.InitHandlers(formats)
		srv.Input()
		addr := srv.addr + ":" + srv.port
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Println(err)
			delete(servers, key)
			continue
		}
		client := irc.NewClient(conn, srv.conf)
		//go client.Run()
		client.Run()
	}
	//ControlLoop(servers)
}
