package main

import (
	"fmt"
	"os"
	"path"

	"github.com/lrstanley/girc"
	"github.com/ubqt-systems/cleanmark"
)

type message struct {
	Name string
	Data string
}

// Append formatted messages to client's buffer string
func (st *State) writeServer(c *girc.Client, e girc.Event) {
	st.event <- []byte("main\n")
	st.writeFile(c, e)
	fmt.Println(string(e.Bytes()))
}

//TODO: Create a struct here, load it with our data, then execute against our string.
//TODO: Clean input for use with clients - unstring markdown, etc.
//TODO: Highlights
//TODO: Update tabs
/* - tabs layout should be as follows
We have our client.channels
map of regular channels
map of highlight channels
(Maybe map, or just uniquely keyed arrays)
*/
func (st *State) writeChannel(c *girc.Client, e girc.Event) {
	st.event <- []byte("main\n")
	st.writeFile(c, e)
	fmt.Println(string(e.Bytes()))
}

func (st *State) writeFile(c *girc.Client, e girc.Event) {
	if e.Params == nil {
		return
	}
	filePath := path.Join(*inPath, c.Config.Server, e.Params[0])
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		fmt.Printf("err %s", err)
		return
	}
	m := &message{Name: e.Source.Name, Data: cleanmark.CleanString(e.Trailing)}
	err = st.chanFmt.Execute(f, m)
	if err != nil {
		fmt.Printf("err %s", err)
		return
	}
	f.WriteString("\n")
}
