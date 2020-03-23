package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/altid/libs/config"
	"github.com/altid/libs/fs"
	"gopkg.in/irc.v3"
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

	conf, err := config.New(buildConfig, *srv, false)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &server{
		cancel: cancel,
	}

	s.parse(conf)

	ctrl, err := fs.CreateCtlFile(ctx, s, conf.Log(), *mtpt, *srv, "feed", *debug)
	if err != nil {
		log.Fatal(err)
	}

	ctrl.CreateBuffer("server", "feed")

	defer ctrl.Cleanup()
	go ctrl.Listen()
	go s.fileListener(ctx, ctrl)

	if e := s.connect(ctx); e != nil {
		log.Fatal(e)
	}

	defer s.conn.Close()

	client := irc.NewClient(s.conn, s.conf)
	client.RunContext(ctx)
}
