package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/lrstanley/girc"
	"github.com/ubqt-systems/ubqtlib"
)

var (
	addr    = flag.String("a", ":4567", "Port to listen on")
	conf    = flag.String("c", "irc.ini", "Configuration file")
	inPath  = flag.String("p", "~/irc", "Path for file system")
	debug   = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

// Client - holds a connected client
type Client struct {
	server  string
	channel string
}

// State - holds server session
type State struct {
	sync.Mutex
	clients map[string]*Client
	irc     map[string]*girc.Client
	tablist []byte
	input   []byte
	event   chan []byte
}

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
	switch filename {
	case "input":
		return st.input, nil
	case "ctl":
		return []byte("part\njoin\nquit\nbuffer\nignore\n"), nil
	case "tabs":
		return st.tablist, nil
	case "status":
		buf, err = st.status(client)
	case "sidebar":
		buf, err = st.sidebar(client)
	case "title":
		buf, err = st.title(client)
	case "feed":
		filePath := path.Join(*inPath, filename)
		if _, err := os.Stat(filePath); err != nil {
			return nil, err
		}
		buf, err = ioutil.ReadFile(filePath)
	default:
		err = errors.New("permission denied")
	}
	return
}

// ClientConnect - add last server in list, first channel in list
func (st *State) ClientConnect(client string) {
	defs := st.clients["default"]
	st.clients[client] = &Client{server: defs.server, channel: defs.channel}
}

// ClientDisconnect - called when client disconnects
func (st *State) ClientDisconnect(client string) {
	delete(st.clients, client)
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st := &State{}
	st.clients = make(map[string]*Client)
	st.irc = make(map[string]*girc.Client)
	st.event = make(chan []byte)
	srv := ubqtlib.NewSrv()
	go func() {
		for {
			select {
			case buf := <-st.event:
				srv.SendEvent(buf)
			}
		}
	}()
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
