package main

/*
Satisfy our fakefile interface for each file type that needs, we may not even need a struct for each in fact
use the existing stat/dir creation, for each true key in state.show as we get messages, run each read, launch in a go routine
*/
import (
	"flag"
	"fmt"
	"os"

	"github.com/thoj/go-ircevent"
)

var (
	addr    = flag.String("a", ":4567", "Port to listen on")
	inPath  = flag.String("p", "~/irc", "Path for file system")
	debug   = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

// State - holds server session
type State struct {
	show   map[string]bool
	input  chan string
	irc    map[string]*irc.Connection
	buffer string
	server string
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st := new(State)
	err := st.Initialize()
	if err != nil {
		fmt.Printf("Err %s", err)
		os.Exit(1)
	}
	// Update data, write to IRC
	go inputHandler(st)
	err = st.Run()
	if err != nil {
		fmt.Printf("Err %s", err)
		os.Exit(1)
	}
}
