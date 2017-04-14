package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sync"
	"text/template"

	"github.com/lrstanley/girc"
	"github.com/ubqt-systems/ubqtlib"
)

var (
	addr    = flag.String("a", ":4567", "Port to listen on")
	conf    = flag.String("c", "irc.ini", "Configuration file")
	inPath  = flag.String("p", path.Join(os.Getenv("HOME"), "irc"), "Path for file system")
	debug   = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

// State - holds server session
type State struct {
	sync.Mutex
	clients map[string]*Client
	irc     map[string]*girc.Client
	tablist map[string]string
	input   []byte
	event   chan []byte
	chanFmt *template.Template
	selfFmt *template.Template
	ntfyFmt *template.Template
	servFmt *template.Template
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st := newState()
	srv := ubqtlib.NewSrv()
	go listenEvents(st, srv)
	if *debug {
		srv.Debug()
	}
	if *verbose {
		srv.Verbose()
	}
	err := st.initialize(srv)
	if err != nil {
		fmt.Printf("Err %s", err)
		os.Exit(1)
	}
	err = srv.Loop(st)
	if err != nil {
		fmt.Printf("Err %s", err)
		os.Exit(1)
	}
}
