package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"aqwari.net/net/styx"
	"github.com/thoj/go-ircevent"
	"github.com/vaughan0/go-ini"
)

var (
	addr    = flag.String("a", ":4567", "Port to listen on")
	debug   = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon := make([]irc.Connection, 1)
	var d = new(Directory)
	for section, _ := range conf {
		if section == "options" {
			//d = setupFiles(conf, section)
			continue
		}
		irccon = append(irccon, *setupServer(conf, section))
	}
	var styxServer styx.Server
	if *verbose {
		styxServer.ErrorLog = log.New(os.Stderr, "", 0)
	}
	if *debug {
		styxServer.TraceLog = log.New(os.Stderr, "", 0)
	}
	styxServer.Addr = *addr
	styxServer.Handler = d

	log.Fatal(styxServer.ListenAndServe())
}
