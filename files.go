package main

import (
	"os"
	"path"
	"fmt"
	"github.com/lrstanley/girc"
	"github.com/ubqt-systems/cleanmark"
)

type message struct {
	Name string
	Data string
}

func (st *State) join(c *girc.Client, e girc.Event) {
	// Make sure our directory exists.
	buffer := path.Join(*inPath, c.Config.Server, e.Params[0])
	err := os.MkdirAll(buffer, 0666)
	if err != nil {
		// Update status to reflect path failure - shouldn't happen
	}
}

// FPRINTF all the things to file.
// TODO: All of these must send a related event so we can update our clients
func (st *State) writeFeed(c *girc.Client, e girc.Event) {
	if e.Params == nil {
		return
	}
	var filePath string
	m = &message{Name: e.Params[0]}
	format := st.chanfmt
	switch e.Command {
	case "ACTION":
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		defer f.Close()
		if err != nil {
			fmt.Printf("err %s", err)
			return
		}
		m.Name = " * "
		m.Data = e.StripAction()
		err = format.Execute(f, m)
		if err != nil {
			fmt.Printf("err %s", err)
		}
		return
	case "NOTICE":
		if m.Name == "ChanServ" {
			m.Name = e.Params[1]
		} else {
			m.Name = c.Config.Server
		}
		st.updateTabs(m.Name, false)
		filePath = path.Join(*inPath, c.Config.Server, m.Name)
	case "MODE":
		m.Name = c.Config.Server
		filePath = path.Join(*inPath, c.Config.Server, m.Name)
	case "PRIVMSG":
		nick := c.GetNick()
		if e.IsFromUser() {
			m.Name = "~" + e.Source.Name
			st.updateTabs(m.Name, true)
			filePath = path.Join(*inPath, c.Config.Server, m.Name)
		}
		if e.IsFromChannel() {
			st.event <- []byte("feed\n")
			filePath = path.Join(*inPath, c.Config.Server, e.Params[0])
			if strings.Contains(e.Trailing, nick) {
				st.updateTabs(m.Name, true)
				format = st.highFmt
			} else {
				st.updateTabs(m.Name, false)
			}
			m.Name = e.Source.Name
		}
	}
	fmt.Println(string(e.Bytes()))
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
defer f.Close()
	if err != nil {
		fmt.Printf("err %s", err)
		return
	}
	m.Data = cleanmark.CleanString(e.Trailing) + "\n"
	m.Name = cleanmark.CleanString(m.Name)
	if e.IsAction() {
		m.Data = cleanmark.CleanString(e.StripAction() + "\n")
		m.Name = " \\* " + e.Source.Name
	}
	err = format.Execute(f, m)
	if err != nil {
		fmt.Printf("err %s", err)
		return
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
