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

func (st *State) updateTabs(name string, hl bool) {
	// Unconditionally add highlights
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
		st.tablist[name] = "[#222222]"
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
	var string filePath
	m := &message{Name: e.Params[0]}
	switch e.Command {
		//case "ACTION":
		//	girc has a pretty print for actions, use that.
		case "NOTICE":
			if m.Name == "ChanServ" {
				m.Name = e.Params[1]
				st.event <- []byte("tabs\n")
			} else {
//TODO: Things like path and who the message are from are different, so hold in two variables instead.
				m.Name = "server"
			}
			filePath = path.Join(*inPath, c.Config.Server, m.Name)
		//	will have to seperate out between server, and chanserv stuff.
		//  like #go-nuts motd thing vs freenode messages
		case "MODE":
			m.Name = "server"
			filePath = path.Join(*inPath, c.Config.Server, m.Name)
		case "PRIVMSG":
			if m.Name == c.GetNick() {
				m.Name = "~" + e.Source.Name
				st.event <- []byte("tabs\n")
				filePath = path.Join(*inPath, c.Config.Server, m.Name)
			} else {
				
			}
	}
	fmt.Println(string(e.Bytes()))
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		fmt.Printf("err %s", err)
		return
	}
	//TODO: Break this out into per-case basis. May have diff info each time?
	m.Data = cleanmark.CleanString(e.Trailing) + "\n"
	err = st.chanFmt.Execute(f, m)
	if err != nil {
		fmt.Printf("err %s", err)
		return
	}
	// Update tabs -- must move out to case
	if strings.Contains(e.Trailing, c.GetNick()) {
		st.updateTabs(m.Name, true)
	} else {
		st.updateTabs(m.Name, false)
	}		
}
