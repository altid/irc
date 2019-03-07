package main

import (
	"flag"
	"log"
	"os"

	"github.com/go-irc/irc"
	fs "github.com/ubqt-systems/fslib"
)

type messageType int
const (
	chanMsg messageType = iota
	selfMsg
	serverMsg
	actionMsg
	none
)

// var match = regexp.MustCompile("([&#][^\\s\\x2C\\x07]{1,199})")

var (
	mtpt = flag.String("p", "/tmp/ubqt", "Path for filesystem (Default /tmp/ubqt)")
	srv  = flag.String("s", "irc.freenode.net", "Name of server to connect to (Default irc.freenode.net)")
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
	ctrl, err := fs.CreateCtrlFile(s, config.log, *mtpt, config.addr, "feed")
	defer ctrl.Cleanup()
	if err != nil {
		log.Fatal(err)
	}
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
