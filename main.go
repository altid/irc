package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/halfwit/ubqt-lib"
	"github.com/vaughan0/go-ini"
	"github.com/thoj/go-ircevent"
)

var (
	debug = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

type Srv map[string]interface{}

func main() {
	//TODO: Just update our struct, irccon member and such.
	flag.Parse()
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		fmt.Printf("Err %s", err)
		os.Exit(1)
	}
	irccon := make([]irc.Connection, 1)
	for section, _ := range conf {
		if section == "options" {
			ubqt.Setup(conf, section)
			continue
		}
		irccon = append(irccon, *setupServer(conf, section))
	}
	srv := new(Srv)
	ubqt.Run(srv)
}
