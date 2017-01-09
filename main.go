package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"aqwari.net/net/styx"
	"github.com/thoj/go-ircevent"
	"github.com/vaughan0/go-ini"
)

var (
	addr    = flag.String("a", ":4567", "Port to listen on")
	inPath  = flag.String("p", "~/irc", "Path for file system")
	debug   = flag.Bool("d", false, "Enable debugging output")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

type State struct {
	file map[string]interface{}
	irc map[string]*irc.Connection
	server     string //current server
	current    string //current buffer
	Title      bool
	Tabs       bool
	Status     bool
	Input      bool //You may want to watch a chat only, for instance
	Sidebar    bool
	Timestamps bool
	ch chan irc.Event
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	var state State
	state.ch = make(chan irc.Event)
	for section, _ := range conf {
		if section == "options" {
			setupState(conf, section, &state)
			continue
		}
		state.server = section
		state.current = section
		setupServer(conf, section, &state)
	}
	state.file = make(map[string]interface{})
	var styxServer styx.Server
	if *verbose {
		styxServer.ErrorLog = log.New(os.Stderr, "", 0)
	}
	if *debug {
		styxServer.TraceLog = log.New(os.Stderr, "", 0)
	}
	styxServer.Addr = *addr
	styxServer.Handler = &state
	log.Fatal(styxServer.ListenAndServe())
}

func walkTo(v interface{}, loc string) (interface{}, bool) {
	cwd := v
	parts := strings.FieldsFunc(loc, func(r rune) bool { return r == '/' })

	for _, p := range parts {
		switch v := cwd.(type) {
		case map[string]interface{}:
			child, ok := v[p]
			if !ok {
				return nil, false
			} else {
				cwd = child
			}
		default:
			return nil, false
		}
	}
	return cwd, true
}

func (st *State) Serve9P(s *styx.Session) {
	var state State
	newState(&state, st)
	go func() {
		event := <-st.ch
		ircobj := st.irc["freenode"]
		/*TODO: We will listen on multiple channels here
		event --> if from current, update state.file["main"]
		if from other, update state.file["tabs"] 
		 - if not already on list & highlighted || if not on list
		input --> write message to server by name 
		*/
		ircobj.Privmsg("#ubqt", "test")
		fmt.Println(event.Arguments[0])
		fmt.Println(event.Arguments[1])
	}()
	for s.Next() {
		t := s.Request()
		file, ok := walkTo(state.file, t.Path())
		if !ok {
			t.Rerror("no such file or directory")
			continue
		}
		fi := &stat{name: path.Base(t.Path()), file: &fakefile{v: file}}
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(fi, nil)
		case styx.Topen:
			switch v := file.(type) {
			case map[string]interface{}:
				t.Ropen(mkdir(v), nil)
			default:
				if s.Request().Path() == "ctl" {
					fmt.Printf("ctl")
				}
				t.Ropen(strings.NewReader(fmt.Sprint(v)), nil)
			}
		case styx.Tstat:
			t.Rstat(fi, nil)
		case styx.Tcreate:
			switch v := file.(type) {
			case map[string]interface{}:
				if t.Mode.IsDir() {
					dir := make(map[string]interface{})
					v[t.Name] = dir
					t.Rcreate(mkdir(dir), nil)
				} else {
					v[t.Name] = new(bytes.Buffer)
					t.Rcreate(&fakefile{
						v:   v[t.Name],
						set: func(s string) { v[t.Name] = s },
					}, nil)
				}
			default:
				t.Rerror("%s is not a directory", t.Path())
			}
		}
	}
}
