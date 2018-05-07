package main

import (
	"os"
	"strings"
	"path"
	"fmt"
	"text/template"
	"github.com/lrstanley/girc"
	//"github.com/ubqt-systems/cleanmark"
)

type message struct {
	Name string
	Data string
}

func (st *State) join(c *girc.Client, e girc.Event) {
	// Make sure our directory exists.
	buffer := path.Join(*inPath, c.Config.Server, e.Params[0])
	err := os.MkdirAll(buffer, 0777)
	if err != nil {
		// Update status to reflect path failure - shouldn't happen
	}
}

func writeFile(m *message, fp string, format *template.Template) {
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = format.Execute(f, m)
	if err != nil {
		// TODO: Log failure
		fmt.Println(err)
	}
	fmt.Fprint(f, "\n")
}

// TODO: Need to create privmsg dirs for pm's
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
		filePath := path.Join(*inPath, c.Config.Server, "server", "feed")
		writeFile(&message{Name: name, Data: e.Trailing}, filePath, st.ntfyFmt) 
	case "MODE":
		filePath := path.Join(*inPath, c.Config.Server, "server", "feed")
		writeFile(&message{Name: c.Config.Server, Data: e.Trailing}, filePath, st.chanFmt)
	case "PRIVMSG":
		name := e.Params[0]
		format := st.chanFmt
		data := e.Trailing
		nick := c.GetNick()
		filePath := path.Join(*inPath, c.Config.Server, e.Params[0], "feed")
		if e.IsFromUser() {
			// Assure we create the directory
			dir := path.Join(*inPath, c.Config.Server, "~" + e.Source.Name)
			filePath = path.Join(dir, "feed")
			os.MkdirAll(dir, 0777)
		}
		if e.IsFromChannel() {
			if strings.Contains(e.Trailing, nick) {
				format = st.highFmt
			}
			name = e.Source.Name
		}
		// TODO: Test if we're at an action here and update `name` accordingly.
		if e.IsAction() {
			format = st.actiFmt
			data = e.StripAction() 
		}
		writeFile(&message{Name: name, Data: data}, filePath, format)
	}
}

// Run through formatter and output to irc.freenode.net/server for example 
func (st *State) writeServer(c *girc.Client, e girc.Event) {
	fmt.Println(e.String())
}

// Remove watch
func (st *State) closeFeed(c *girc.Client, e girc.Event) {
	fmt.Println(e.String())
}

// Log to feed as well as update `status` when it relates to user
func (st *State) mode(c *girc.Client, e girc.Event) {
	fmt.Println(e.String())
}

// Remove all watches
func (st *State) quitServer(c *girc.Client, e girc.Event) {}

// Log to channel and update out `title`
func (st *State) topic(c *girc.Client, e girc.Event) {
	fmt.Println(e.String())
}

func (st *State) InLoop() {
	<-st.done
}
