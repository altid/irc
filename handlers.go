package main

import (
	"bytes"
	"encoding/gob"
)

// handleInput - append valid runes to input type, curtail input at [history]input lines.
func (st *State) handleInput(data []byte, name string) (int, error) {
	// TODO: Scrub and send message to channel
	st.input = append(st.input, data...)
	return len(data), nil
}

func (st *State) handleCtl(data []byte, name string) (int, error) {
	//TODO: Handle command
	return len(data), nil
}

// st.irc[client] will give us the active struct for the channel our client is currently displaying
// from there we can update status, title, sidebar, main accordingly; as well as take any input to update tabs.
func (st *State) ctl(client string) ([]byte, error) {
	return []byte("part\njoin\nquit\nbuffer\n"), nil
}

func (st *State) status(client string) ([]byte, error) {
	//Channel.Modes //Channel.Name
	return []byte("status file\n"), nil
}

func (st *State) sidebar(client string) ([]byte, error) {
	channel := st.channel[client]
	list := channel.NickList() //(returns []string)
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(list)
	return buf.Bytes(), nil
	//return []byte("sidebar file\n"), nil
}

func (st *State) tabs(client string) ([]byte, error) {
	//return(st.tabs)
	return []byte("tabs file\n"), nil
}

func (st *State) buff(client string) ([]byte, error) {
	//open file and such
	return []byte("buffer file\n"), nil
}

func (st *State) title(client string) ([]byte, error) {
	return []byte("title file\n"), nil
}
