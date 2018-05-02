package main

import (
	"fmt"
	"github.com/lrstanley/girc"
)
/* 
[#color](.Name) .Data
event.Source.Name  // nickname/server/service
event.Source.Ident // 'user'
event.Trailing // Data
event.Timestamp
Client.Config.Server // name of our server
*/

// TODO: All of these must send a related event so we can update our clients
func (st *State) writeFeed(c *girc.Client, e girc.Event) {
	fmt.Println(e.String())
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
