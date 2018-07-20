package main

// TODO: Switch to events loop for adding/removing directories and associated files

import (
	"os"
	"io/ioutil"
	"log"
	"strings"
	"path"
	"fmt"
	"time"
	"text/template"
	"github.com/lrstanley/girc"
	"github.com/ubqt-systems/cleanmark"
)

type message struct {
	Name string
	Data string
}

func SendEvent(data string) {
	f, err := os.OpenFile(path.Join(*inPath, "events"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Print(err)
	}
	defer f.Close()
	fmt.Fprintln(f, data)
}

func (st *State) join(c *girc.Client, e girc.Event) {
	// TODO - Smart Filters: Add other user to map[username]timestamp, for Smart filters
	if c.GetNick() == e.Source.Name {
		go InitBuffer(c, e)
	} else {
		SendEvent(path.Join(c.Config.Server, e.Params[0], "feed"))
	}
}

func InitBuffer(c *girc.Client, e girc.Event) {
	// TODO: Have a timed block that waits til we have a non-empty string for each value
	// This sleep is a necessity to let the data propagate until then
	time.Sleep(time.Second * 3)
	joined := c.LookupChannel(e.Params[0])
	filePath := path.Join(*inPath, c.Config.Server, e.Params[0])
	data := cleanmark.CleanString(e.Source.Name + " (" + joined.Modes.String() + ")")
	ioutil.WriteFile(path.Join(filePath, "status"), []byte(data), 0666)
	data = cleanmark.CleanString(joined.Topic)
	ioutil.WriteFile(path.Join(filePath, "title"), []byte(data), 0666)
	//joined.Trusted(c) - list of members who are voiced
	//joined.Admins(c) - list of members who are at least halfops
	//joined.UserList - list of all members
	// TODO: Loop through joined.Admins and joined.Trusted, add them to the file at the top - and remove the named entries from UserList
	// Finally, add all entries left from UserList
	// Send events for all files updated here
	SendEvent(path.Join(c.Config.Server, e.Params[0], "status"))
	SendEvent(path.Join(c.Config.Server, e.Params[0], "title"))
	SendEvent(path.Join(c.Config.Server, e.Params[0], "feed"))
	//SendEvent(path.Join(c.Config.Server, e.Params[0], "sidebar"))
	//SendEvent(path.Join(c.Config.Server, e.Params[0], "tabs"))
}

func writeFile(m *message, fp string, format *template.Template) {
	f, err := os.OpenFile(path.Join(*inPath, fp), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	m.Name = cleanmark.CleanString(m.Name)
	m.Data = cleanmark.CleanString(m.Data)
	err = format.Execute(f, m)
	if err != nil {
		// TODO: Log failure
		fmt.Println(err)
	}
	fmt.Fprint(f, "\n")
	SendEvent(fp)
}

func (st *State) writeFeed(c *girc.Client, e girc.Event) {
	if e.Params == nil {
		return
	}
	switch e.Command {
	case "NOTICE":
		var name string
		if e.Params[0] == "ChanServ" {
			name = e.Params[1]
		} else {
			name = c.Config.Server
		}
		filePath := path.Join(c.Config.Server, "server", "feed")
		writeFile(&message{Name: name, Data: e.Trailing}, filePath, st.cfg.ntfyFmt) 
	case "MODE":
		filePath := path.Join(c.Config.Server, "server", "feed")
		writeFile(&message{Name: c.Config.Server, Data: e.Trailing}, filePath, st.cfg.chanFmt)
	case "PRIVMSG":
		name := e.Params[0]
		format := st.cfg.chanFmt
		data := e.Trailing
		nick := c.GetNick()
		filePath := path.Join(c.Config.Server, e.Params[0], "feed")
		if e.IsFromUser() {
			// Assure we create the directory
			name = e.Source.Name
			dir := path.Join(c.Config.Server, "~" + e.Source.Name)
			filePath = path.Join(dir, "feed")
			os.MkdirAll(path.Join(*inPath, dir), 0777)
		}
		if e.IsFromChannel() {
			if strings.Contains(e.Trailing, nick) {
				format = st.cfg.highFmt
			}
			name = e.Source.Name
		}
		// TODO: Test if we're at an action here and update `name` accordingly.
		if e.IsAction() {
			format = st.cfg.actiFmt
			data = e.StripAction() 
		}
		writeFile(&message{Name: name, Data: data}, filePath, format)
	}
}

// Run through formatter and output to irc.freenode.net/server for example 
func (st *State) writeServer(c *girc.Client, e girc.Event) {
	filePath := path.Join(c.Config.Server, "server", "feed")
	writeFile(&message{Name: "Server", Data: e.Trailing}, filePath, st.cfg.chanFmt) 
}

// Remove watch
func (st *State) part(c *girc.Client, e girc.Event) {
	// TODO - Cleanup files: If this is us parting, remove title/status/unregister event, etc
}

// Log to feed as well as update `status` when it relates to user
func (st *State) mode(c *girc.Client, e girc.Event) {
	// Output to status with current channel, mode, etc
	parnum := len(e.Params)
	if parnum == 1 {
		return
	}
	filePath := path.Join(c.Config.Server, e.Params[0])
	writeFile(&message{Name: e.Source.Name, Data: strings.Join(e.Params[1:], " ")}, path.Join(filePath, "feed"), st.cfg.modeFmt)
}

// Remove all watches
func (st *State) quitServer(c *girc.Client, e girc.Event) {
	// TODO: close all threads and delete all but feed file
}

// Log to channel and update `title` file.
func (st *State) topic(c *girc.Client, e girc.Event) {
	filePath := path.Join(c.Config.Server, e.Params[0])
	writeFile(&message{Name: e.Source.Name, Data: "has changed the topic to \"" + e.Trailing + "\""}, path.Join(filePath, "feed"), st.cfg.chanFmt) 
	data := cleanmark.CleanString(e.Trailing)
	ioutil.WriteFile(path.Join(*inPath, filePath, "title"), []byte(data), 0666)
}
