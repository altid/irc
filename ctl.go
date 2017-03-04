package main

import (
	"fmt"

	"github.com/lrstanley/girc"
)

func (st *State) handleSet(b [][]byte, client string) {

}

func (st *State) handleMsg(nick string, message string, client string) {
	if !girc.IsValidNick(nick) {
		return
	}
	current := st.clients[client]
	irc := st.irc[current.server]
	irc.Commands.Message(nick, message)
}

func (st *State) handleJoin(channel string, client string) {
	//TODO: if string(b[1]) == "-server" {
	if !girc.IsValidChannel(channel) {
		return
	}
	server := st.irc[st.clients[client].server]
	err := server.Commands.Join(channel)
	if err != nil {
		fmt.Println("Join failed")
		return
	}
	st.clients[client].channel = channel
}

func (st *State) handlePart(channel string, client string) {
	server := st.irc[st.clients[client].server]
	err := server.Commands.Part(channel, "")
	if err != nil {
		fmt.Println("Part failed")
		return
	}
	st.clients[client].channel = server.Channels()[0]
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
