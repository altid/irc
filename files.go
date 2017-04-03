package main

import (
	"fmt"
	//"text/template"

	"github.com/lrstanley/girc"
)

// Append formatted messages to client's buffer string
func (st *State) writeServer(c *girc.Client, e girc.Event) {
	fmt.Println(string(e.Bytes()))
}

func (st *State) writeChannel(c *girc.Client, e girc.Event) {
	st.event <- []byte("main\n")
	fmt.Println(string(e.Bytes()))
}
