package main

import (
	"flag"
	"log"
	"os"

	"github.com/altid/libs/fs"
	"github.com/go-irc/irc"
)

var (
	mtpt = flag.String("p", "/tmp/altid", "Path for filesystem")
	srv  = flag.String("s", "irc", "Name of service")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}

	config, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	s := newServer(config)
	ctrl, err := fs.CreateCtlFile(s, config.log, *mtpt, *srv, "feed")
	if err != nil {
		log.Fatal(err)
	}

	defer ctrl.Cleanup()

	// Make a type which never will log
	ctrl.CreateBuffer("server", "feed")
	ctx, err := ctrl.Start()
	if err != nil {
		log.Fatal(err)
	}

	go s.fileListener(ctx, ctrl)
	err = s.connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client := irc.NewClient(s.conn, s.conf)
	client.RunContext(ctx)
}
