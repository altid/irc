package main

import (
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
func (st *State) ClientWrite(filename string, client string, data []byte) (int, error) {
	// TODO: If struct doesn't exist, create one
	if filename == "input" {
		st.input = append(st.input, data...)
	}
	return len(data), nil
}

// ClientRead - Return formatted strings for various files
func (st *State) ClientRead(filename string, client string) ([]byte, error) {
	if filename == "input" {
		return st.input, nil
	}
	return []byte("Hello world\n"), nil
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
