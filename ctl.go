package main

import "fmt"

func (st *State) handleSet(b [][]byte, client string) {

}

func (st *State) handleMsg(b [][]byte, client string) {

}

func (st *State) handleJoin(channel string, client string) {
	//if string(b[1]) == "-server" {
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

func (st *State) handleIgnore(b [][]byte, client string) {

}
