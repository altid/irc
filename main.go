package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ubqt-systems/ubqtlib"
)

var (
	addr    = flag.String("a", ":4567", "Port to listen on")
	inPath  = flag.String("p", "~/irc", "Path for file system")
	debug   = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

// State - holds server session
type State struct {
	buffer string
	server string
	input  []byte
}

var clients map[string]*State

// ClientWrite - Handle writes on ctl, input to send to channel/mutate program state
func (st *State) ClientWrite(filename string, client string, data []byte) (n int, err error) {
	switch filename {
	case "input":
		n, err = st.handleInput(data, client)
	case "ctl":
		n, err = st.handleCtl(data, client)
	default:
		err = errors.New("permission denied")
	}
	return
}

// ClientRead - Return formatted strings for various files
func (st *State) ClientRead(filename string, client string) (buf []byte, err error) {
	// Calls may error, pass that back as required
	switch filename {
	case "input":
		return st.input, nil
	case "ctl":
		buf, err = st.ctl(client)
	case "status":
		buf, err = st.status(client)
	case "sidebar":
		buf, err = st.sidebar(client)
	case "tabs":
		buf, err = st.tabs(client)
	case "main":
		buf, err = st.buff(client)
	case "title":
		buf, err = st.title(client)
	default:
		err = errors.New("permission denied")
	}
	return
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st := &State{}
	clients = make(map[string]*State)
	srv := ubqtlib.NewSrv()
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
