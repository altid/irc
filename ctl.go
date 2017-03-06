package main

import (
	"fmt"

	"github.com/lrstanley/girc"
)

func (st *State) handleSet(b [][]byte, client string) {
	// Toggle off/on UI elements
	// Toggle timestapms
}

func (st *State) handleMsg(nick string, message string, client string) {
	if !girc.IsValidNick(nick) {
		return
	}
	current := st.clients[client]
	irc := st.irc[current.server]
	irc.Commands.Message(nick, message)
}

//TODO: Handle per-server as well as passwords.
func (st *State) handleJoin(channel string, client string) {
	if !girc.IsValidChannel(channel) {
		return
	}
	current := st.clients[client]
	irc := st.irc[current.server]
	err := irc.Commands.Join(channel)
	if err != nil {
		fmt.Println("Join failed")
		return
	}
	current.channel = channel
}

func (st *State) handlePart(channel string, client string) {
	current := st.clients[client]
	irc := st.irc[current.server]
	err := irc.Commands.Part(channel, "leaving")
	if err != nil {
		fmt.Println("Part failed")
		return
	}
	current.channel = irc.Channels()[0]
}

// TODO: Handle cases where we swap to a buffer
// on another network - range st.c and test if buffer exists
func (st *State) handleBuffer(channel string, client string) {
	st.clients[client].channel = channel
}

// TODO: Store a hardcoded ignore list that we source on startup, handle here
func (st *State) handleIgnore(b [][]byte, client string) {
	switch string(b[0]) {
	case "add":
	case "del":
	case "list":
	}
}
