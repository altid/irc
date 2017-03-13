package main

import (
	"fmt"
	//"text/template"

	"github.com/lrstanley/girc"
)

func (st *State) writeServer(c *girc.Client, e girc.Event) {
	fmt.Println(string(e.Bytes()))
}

func (st *State) writeChannel(c *girc.Client, e girc.Event) {
	go st.wait("main")
	fmt.Println(string(e.Bytes()))
}

func (st *State) wait(title string) {
	switch title {
	case "title":
		<-st.titlch
	case "main":
		<-st.buffch
	case "tabs":
		<-st.tabsch
	case "status":
		<-st.statch
	case "sidebar":
		<-st.sidech
	}
}
