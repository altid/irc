package main

import (
	"flag"
	"log"
	"os"

	"github.com/go-irc/irc"
	fs "github.com/ubqt-systems/fslib"
)

var (
	mtpt = flag.String("p", "/tmp/ubqt", "Path for filesystem (Default /tmp/ubqt)")
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
	ctrl, err := fs.CreateCtrlFile(s, config.log, *mtpt, *srv, "feed")
	defer ctrl.Cleanup()
	if err != nil {
		log.Fatal(err)
	}
	// Make a type which never will log
	//ctrl.CreateTempBuffer("server", "feed")
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
