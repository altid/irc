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

type state struct {
	file       map[string]interface{}
	irc        map[string]*irc.Connection
	server     string
	current    string //current buffer
	Title      bool
	Tabs       bool
	Status     bool
	Input      bool //You may want to watch a chat only, for instance
	Sidebar    bool
	Timestamps bool
	ch         chan irc.Event
	input      chan string
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
	var st state
	st.ch = make(chan irc.Event)
	for section := range conf {
		if section == "options" {
			setupState(conf, section, &st)
			continue
		}
		st.server = section
		st.current = section
		setupServer(conf, section, &st)
	}
	st.file = make(map[string]interface{})
	var styxServer styx.Server
	if *verbose {
		styxServer.ErrorLog = log.New(os.Stderr, "", 0)
	}
	if *debug {
		styxServer.TraceLog = log.New(os.Stderr, "", 0)
	}
	styxServer.Addr = *addr
	styxServer.Handler = &st
	go func() {
		select {
		case event := <-st.ch:
			fmt.Println(event)
		case input := <-st.input:
			ircobj := st.irc[st.server]
			ircobj.Privmsg(st.current, input)
		}
	}()
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
			}
			cwd = child
		default:
			return nil, false
		}
	}
	return cwd, true
}

//TODO: Starting from here, blow up everything styx side. Reimplement everything and see what we can define for our own data type.
// No reason we need to use the naive interface that the example does, we can attribute whatever names and such are necessary to make this all work
func (st *state) Serve9P(s *styx.Session) {
	var client state
	newState(&client, st)
	for s.Next() {
		t := s.Request()
		name := path.Base(t.Path())
		file, ok := walkTo(client.file, t.Path())
		if !ok {
			t.Rerror("no such file or directory")
			continue
		}
		fi := &stat{name: name, file: &fakefile{v: file}}
		switch t := t.(type) {
		case styx.Twalk:
			t.Rwalk(fi, nil)
		case styx.Topen:
			switch v := file.(type) {
			case map[string]interface{}:
				t.Ropen(mkdir(v), nil)
			default:
				//TODO: This updates after an iteration, oddly.
				//TODO: Updating our data here will likely not be very useful
				if name == "input" {
					client.file["input"] = "Hey, we got some input"
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
