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

//TODO: Switch to lstanley/girc instead, it's still event based but much more mature, with better tracking

// State - holds server session
type State struct {
	buffer string
	server string
}

// WriteFile - Handle writes on ctl, input to send to channel/mutate program state
func (st *State) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return nil
}

// ReadFile - Return formatted strings for various files
func (st *State) ReadFile(filename string) ([]byte, error) {
	return []byte("Hello world\n"), nil
}

// CloseFile - Remove file from our working list (perclient)
func (st *State) CloseFile(filename string) error {
	return nil
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st := &State{}
	srv := ubqtlib.NewSrv()
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
