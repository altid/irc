package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thoj/go-ircevent"
	"github.com/vaughan0/go-ini"
)

var (
	addr    = flag.String("a", ":4567", "Port to listen on")
	inPath  = flag.String("p", "~/irc", "Path for file system")
	debug   = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

type show struct {
	Title bool
	Tabs bool
	Status bool
	Input bool //You may want to watch a chat only, for instance
	Sidebar bool
	Timestamps bool
}

type server struct {
	file map[string]interface{}
}

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
	var srv server
	show := new(show)
	for section, _ := range conf {
		if section == "options" {
			show = setupShow(conf, section, &srv)
			continue
		}
		irccon = append(irccon, *setupServer(conf, section))
	}
	fmt.Printf("Defaults\nShow title: %v\nShow tabs: %v\nShow status: %v\nShow input: %v\nShow sidebar: %v\n", show.Title, show.Tabs, show.Status, show.Input, show.Sidebar)
}
