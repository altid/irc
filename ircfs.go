package main

import (
	"flag"
	"log"
	"os"

	"github.com/altid/libs/config"
	"github.com/altid/libs/fs"
	"github.com/go-irc/irc"
)

var (
	mtpt  = flag.String("p", "/tmp/altid", "Path for filesystem")
	srv   = flag.String("s", "irc", "Name of service")
	debug = flag.Bool("d", false, "enable debug printing")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}

	conf, err := config.New(buildConfig, *srv)
	if err != nil {
		log.Fatal(err)
	}

	s := &server{}
	s.parse(conf)

	ctrl, err := fs.CreateCtlFile(s, conf.Log(), *mtpt, *srv, "feed", *debug)
	if err != nil {
		log.Fatal(err)
	}

	defer ctrl.Cleanup()

	ctrl.CreateBuffer("server", "feed")

	ctx, err := ctrl.Start()
	if err != nil {
		log.Fatal(err)
	}

	go s.fileListener(ctx, ctrl)
	if e := s.connect(ctx); e != nil {
		log.Fatal(e)
	}
	defer s.conn.Close()

	client := irc.NewClient(s.conn, s.conf)
	client.RunContext(ctx)
}
