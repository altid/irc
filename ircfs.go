package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/user"

	"github.com/altid/libs/config"
	"github.com/altid/libs/fs"
	"gopkg.in/irc.v3"
)

var (
	mtpt  = flag.String("p", "/tmp/altid", "Path for filesystem")
	srv   = flag.String("s", "irc", "Name of service")
	debug = flag.Bool("d", false, "enable debug printing")
	setup = flag.Bool("conf", false, "run configuration setup")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}

	u, _ := user.Current()

	conf := &defaults{
		Address: "irc.freenode.net",
		Auth:    "password",
		SSL:     "none",
		Port:    6697,
		Filter:  "none",
		Nick:    u.Name,
		Name:    "guest",
		User:    "guest",
		Logdir:  "none",
		TLSCert: "none",
		TLSKey:  "none",
	}

	if *setup {
		if e := config.Create(conf, *srv, "", *debug); e != nil {
			log.Fatal(e)
		}

		os.Exit(0)
	}

	if e := config.Marshal(conf, *srv, "", *debug); e != nil {
		log.Printf("config file malformed or missing. Please run %s -c or manually repair", os.Args[0])
		log.Fatal(e)
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &server{
		cancel: cancel,
		d:      conf,
	}

	s.parse()

	ctrl, err := fs.CreateCtlFile(ctx, s, string(conf.Logdir), *mtpt, *srv, "feed", *debug)
	if err != nil {
		log.Fatal(err)
	}

	defer ctrl.Cleanup()
	ctrl.CreateBuffer("server", "feed")
	go ctrl.Listen()
	go s.fileListener(ctx, ctrl)

	if e := s.connect(ctx); e != nil {
		log.Fatal(e)
	}

	defer s.conn.Close()
	client := irc.NewClient(s.conn, s.conf)
	client.RunContext(ctx)
}
