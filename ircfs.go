package main

import (
	"flag"
	"log"
	"os"

	"github.com/go-irc/irc"
	fs "github.com/ubqt-systems/fslib"
)

const (
	inputMsg messageType = iota
	chanMsg
	selfMsg
	serverMsg
	notifyMsg
	titleMsg
	statusMsg
	highMsg
	actionMsg
	none
)

type messageType int

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
	srv := newServer(config)
	ctrl, err := fs.CreateCtrlFile(srv, config.log, *mtpt, config.addr, "feed")
	defer ctrl.Cleanup()
	if err != nil {
		log.Fatal(err)
	}
	ctx, err := ctrl.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = srv.connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// Ensure we close when the context from ctrl finishes
	// DialContext only watches context until the connection is good
	go func() {
		<-ctx.Done()
		srv.conn.Close()
	}()
	client := irc.NewClient(srv.conn, srv.conf)
	client.RunContext(ctx)
}
