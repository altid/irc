package main

import (
	"flag"
	"log"
	"os"

	"github.com/altid/ircfs"
)

var (
	srv   = flag.String("s", "irc", "Name of service")
	mdns  = flag.Bool("m", false, "enable mDNS broadcast of service")
	debug = flag.Bool("d", false, "enable debug printing")
	ssh   = flag.Bool("x", false, "enable ssh listener (default is 9p)")
	ldir  = flag.Bool("l", false, "enable logging for main buffers")
	setup = flag.Bool("conf", false, "run configuration setup")
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

	ircfs, err := ircfs.Register(*ssh, *ldir, *srv, *debug)
	if err != nil {
		log.Fatal(err)
	}

	defer ircfs.Cleanup()
	if *mdns {
		ircfs.Broadcast()
	}

	if e := ircfs.Start(); e != nil {
		log.Fatal(e)
	}

	if e := ircfs.Listen(); e != nil {
		log.Fatal(e)
	}
}
