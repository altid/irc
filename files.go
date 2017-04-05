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

/*
p := filepath.Join(*inPath, c.Name, e.Arguments[0])
if e.Arguments[0] == c.User {
	p = filepath.Join(*inPath, c.Name, e.Nick)
open and such
const format = `[#5F87A7]({{index .Arguments 0}}) {{index .Arguments 1}}`
if err != nil {
	fmt.Printf("Err %s", err)
}

err = t.Execute(f, e)
if err != nil {
	fmt.Printf("Err %s", err)
}
*/
