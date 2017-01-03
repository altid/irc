package main

import (
	"bytes"
	"flag"
	"path/filepath"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"

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

type Session struct {
	Title          string
	BufferList     string
	NickList       string
	Status         string
	CompletionList string
	Main           string
	Current        string
	mu             sync.Mutex
}

type Show struct {
	Title      bool
	Tabs       bool
	Status     bool
	Input      bool //You may want to watch a chat only, for instance
	Sidebar    bool
	Timestamps bool
}

type Server struct {
	file map[string]interface{}
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	var srv Server
	conf, err := ini.LoadFile("irc.ini")
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon := make([]irc.Connection, 1)
	Show := new(Show)
	Session := new(Session)
	Session.Current = "freenode/#ubqt"
	for section, _ := range conf {
		if section == "options" {
			Show = setupShow(conf, section)
			continue
		}
		irccon = append(irccon, *setupServer(conf, section, Session))
	}
	srv.Update(Show, Session)
	var styxServer styx.Server
	if *verbose {
		styxServer.ErrorLog = log.New(os.Stderr, "", 0)
	}
	if *debug {
		styxServer.TraceLog = log.New(os.Stderr, "", 0)
	}
	styxServer.Addr = *addr
	styxServer.Handler = &srv
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

func (srv *Server) Serve9P(s *styx.Session) {
	for s.Next() {
		t := s.Request()
		file, ok := walkTo(srv.file, t.Path())
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



func (s Session) Read(file string) string {
	p := filepath.Join(*inPath, file)
	f, err := os.OpenFile(p, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("Err %s", err)
	}
	defer f.Close()
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (s Session) UpdateStatus() string {
	return "status"
}

func (s Session) UpdateTabs() string {
	return "lots of things for the tab bar"
}

func (s Session) UpdateSidebar() string {
	return "items\non\nthe\nside\nare\ncool"
}

func (s Session) ListFunctions() string {
	return "buffer\njoin\npart"
}
