package main

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

//TODO: Switch to lstanley/girc instead, it's still event based but much more mature, with better tracking
//TODO: Abstract back to a library
// ubqt type
// addr, inpath, debug, verbose are used to set up
// ubqt.Start() afterwards
// ubqt.NewSession(), defaults for addr inPath, debug, verbose
// ubqt.ListenAndServe() could be a dummy wrapper too.
// See how to make these available as functions to anything
// that satisfies our interface
// Special filename input, ctl, etc
// type ubqt interface {
//     ReadFile(filename string) ([]byte, error)
//     WriteFile(filename string, data []byte, perm os.FileMode)  error
//}
// Have an enum of our explicit file names, mapped to strings
// select {
// case UB_INPUT
//   - etc
// default:
//   - check if file exists (get path)
// So it'd be implemented as a reader/writer/closer, and then we just handle it
// Maybe ReadFile/WriteFile/CloseFile
// ReadFile(b []byte, f *File)
// Implement functions, 9p handles the rest.
// If no data == nil, don't draw file

// State - holds server session
type State struct {
	show   map[string]bool
	irc    map[string]*irc.Connection
	event  chan string
	done   chan int
	buffer string
	server string
	// This data is shared across all client sessions
	// input for history of entered commands, this is across all channels
	//TODO: input history: shared or per client
	input []byte
	// Tabs will include the current, plus all with activity
	tabs []byte
	ctl  string
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st := &State{}
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
