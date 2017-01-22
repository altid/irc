package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lionkov/go9p/p/srv"
	"github.com/thoj/go-ircevent"
)

var (
	addr    = flag.String("a", ":4567", "Port to listen on")
	inPath  = flag.String("p", "~/irc", "Path for file system")
	debug   = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

type state struct {
	irc map[string]*irc.Connection
	// Represents files served over 9p
	current *Current
	bar     *Sidebar
	input   *Input
	title   *Title
	tabs    *Tabs
	status  *Status
	ctl     *Ctl
	// Toggle for timestamps
	timestamps map[string]bool
}

//TODO: This will be cleaned up in the future, for now get something running

// Current buffer, active on server
type Current struct {
	srv.File
	// Used to have per-client buffers
	server map[string]string
	buffer map[string]string
}

// Sidebar holds a list of nicknames present in current channel
type Sidebar struct {
	show bool
	srv.File
	names map[string][]string
}

// Ctl - List of completions, on read; on write commands
type Ctl struct {
	srv.File
	compl []byte
}

// Title will print the name of our program
type Title struct {
	show bool
	srv.File
}

// Tabs holds list of buffers
type Tabs struct {
	show bool
	srv.File
	buflist []string
}

// Input accepts user input, will scrub for slash commands
type Input struct {
	show bool
	srv.File
	history []byte
	ch      chan []byte
}

// Status lists current user count, channel modes, etc
type Status struct {
	show bool
	srv.File
	// To be changed to explicit status members
	items []string
	mode  []byte
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st, err := newState()
	if err != nil {
		fmt.Printf("Err %s", err)
		os.Exit(1)
	}
	root, err := setupFiles(st)
	if err != nil {
		fmt.Printf("Err %s", err)
		os.Exit(1)
	}
	s := srv.NewFileSrv(root)
	s.Dotu = true
	s.Start(s)

	err = s.StartNetListener("tcp", *addr)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

}
