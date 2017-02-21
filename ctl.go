package main

import "fmt"

func (st *State) handleSet(b [][]byte, client string) {

}

func (st *State) handleMsg(b [][]byte, client string) {

}

func (st *State) handleJoin(channel string, client string) {
	//if string(b[1]) == "-server" {
	server := st.c[client].irc
	err := server.Commands.Join(channel)
	if err != nil {
		fmt.Println("Join failed")
	}
}

func (st *State) handlePart(b [][]byte, client string) {

}

// TODO: Handle cases where we swap to a buffer
// on another network - range st.c and test if buffer exists
func (st *State) handleBuffer(channel string, client string) {
	st.c[client].channel = channel
}

func (st *State) handleIgnore(b [][]byte, client string) {

}
