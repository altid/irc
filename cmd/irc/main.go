package main

import (
	"flag"
	"log"
	"os"

	"github.com/altid/irc"
)

var (
	srv   = flag.String("s", "irc", "name of service")
	debug = flag.Bool("d", false, "enable debug printing")
	setup = flag.Bool("conf", false, "run configuration setup")
	fg = flag.Bool("f", false, "run in the foreground")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}

	if *setup {
		if e := irc.CreateConfig(*srv, *debug); e != nil {
			log.Fatal(e)
		}
		os.Exit(0)
	}

	svc, err := irc.Register(*srv, *fg, *debug)
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Cleanup()
	if e := svc.Run(); e != nil {
		log.Fatal(e)
	}
}

func runmain() error {

	return nil
}
