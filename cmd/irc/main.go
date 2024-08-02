package main

import (
	"flag"
	"log"
	"os"

	"github.com/altid/ircfs"
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
		if e := ircfs.CreateConfig(*srv, *debug); e != nil {
			log.Fatal(e)
		}
		os.Exit(0)
	}

	irc, err := ircfs.Register(*srv, *fg, *debug)
	if err != nil {
		log.Fatal(err)
	}
	defer irc.Cleanup()
	if e := irc.Run(); e != nil {
		log.Fatal(e)
	}
}

func runmain() error {

	return nil
}
