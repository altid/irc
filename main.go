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

type show struct {
	Title bool
	Tabs bool
	Status bool
	Input bool //You may want to watch a chat only, for instance
	Sidebar bool
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
	show := new(show)
	for section, _ := range conf {
		if section == "options" {
			show = setupShow(conf, section, &srv)
			continue
		}
		irccon = append(irccon, *setupServer(conf, section))
	}
	var styxServer styx.Server
	if *verbose {
		styxServer.ErrorLog = log.New(os.Stderr, "",0)
	}
	if *debug {
		styxServer.TraceLog = log.New(os.Stderr, "", 0)
	}
	styxServer.Addr = *addr
	styxServer.Handler = &srv
	fmt.Printf("Defaults\nShow title: %v\nShow tabs: %v\nShow status: %v\nShow input: %v\nShow sidebar: %v\n", show.Title, show.Tabs, show.Status, show.Input, show.Sidebar)
	log.Fatal(styxServer.ListenAndServe())
}

func walkTo(v interface{}, loc string) (interface{}, bool){
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
		switch t:= t.(type) {
		case styx.Twalk:
			t.Rwalk(fi, nil)
		case styx.Topen:
			switch v := file.(type) {
			case map[string]interface{}:
				t.Ropen(mkdir(v), nil)
			default:
				t.Ropen(strings.NewReader(fmt.Sprint(v)), nil)
			}
		case styx.Tstat:
			t.Rstat(fi, nil)
		case styx.Tcreate:
			switch v := file.(type) {
			case map[string]interface{}:
				if t.Mode.IsDir(){
					dir := make(map[string]interface{})
					v[t.Name] = dir
					t.Rcreate(mkdir(dir), nil)
				} else {
					v[t.Name] = new(bytes.Buffer)
					t.Rcreate(&fakefile{
						v: v[t.Name],
						set: func(s string) { v[t.Name] = s},
					}, nil)
				}
			default:
				t.Rerror("%s is not a directory", t.Path())
			}
		}
	}
}
