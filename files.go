package main

import (
	"fmt"
	"os"
	"path"

	"github.com/lrstanley/girc"
)

// Append formatted messages to client's buffer string
func (st *State) writeServer(c *girc.Client, e girc.Event) {
	fmt.Println(string(e.Bytes()))
}

func (st *State) writeChannel(c *girc.Client, e girc.Event) {
	st.event <- []byte("main\n")
	filePath := path.Join(*inPath, c.Config.Server, e.Params[0])
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		fmt.Printf("err %s", err)
		return
	}
	err = st.chanFmt.Execute(f, e)
	if err != nil {
		fmt.Printf("err %s", err)
		return
	}
	f.WriteString("\n")
	fmt.Println(string(e.Bytes()))
}
