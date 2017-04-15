package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/lrstanley/girc"
	"github.com/ubqt-systems/cleanmark"
)

type message struct {
	Name string
	Data string
}

func (st *State) updateTabs(n string, hl bool) {
	// Unconditionally add highlights
	name := cleanmark.CleanString(n)
	if hl {
		st.Lock()
		st.tablist[name] = "[#9d0006]"
		st.Unlock()
		st.event <- []byte("tabs\n")
		return
	}
	// Else, add channel if it doesn't exist (guards highlight overwrites, etc)
	if _, ok := st.tablist[name]; !ok {
		st.Lock()
		st.tablist[name] = "[#928374]"
		st.Unlock()
		st.event <- []byte("tabs\n")
	}
}

//TODO: ACTION
//TODO: MODE
//TODO: Highlights

func (st *State) writeFile(c *girc.Client, e girc.Event) {
	//Source has Name, Ident, Host (don't need host)
	// Start our empty struct, fill it later
	if e.Params == nil {
		return
	}
	var filePath string
	m := &message{Name: e.Params[0]}
	format := st.chanFmt
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
			switch {
				case e.IsFromUser():
					m.Name = "~" + e.Source.Name
					st.updateTabs(m.Name, true)
					filePath = path.Join(*inPath, c.Config.Server, m.Name)
				case e.IsFromChannel():
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
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
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
