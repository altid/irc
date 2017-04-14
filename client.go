package main

import (
	"errors"
	"os"
	"path"
)

// Client - A connected client
type Client struct {
	server  string
	channel string
}

// ClientWrite - Handle writes on ctl, input to send to channel/mutate program state
func (st *State) ClientWrite(filename string, client string, data []byte) (n int, err error) {
	switch filename {
	case "input":
		n, err = st.handleInput(data, client)
	case "ctl":
		n, err = st.handleCtl(data, client)
	default:
		err = errors.New("permission denied")
	}
	return
}

// ClientRead - Return formatted strings for various files
func (st *State) ClientRead(filename string, client string) (buf []byte, err error) {
	switch filename {
	case "input":
		return st.input, nil
	case "ctl":
		return []byte("part\njoin\nquit\nbuffer\nignore\n"), nil
	case "tabs":
		buf, err = st.tabs(client)
	case "status":
		buf, err = st.status(client)
	case "sidebar":
		buf, err = st.sidebar(client)
	case "title":
		buf, err = st.title(client)
	default:
		err = errors.New("permission denied")
	}
	return
}

// ClientOther - Should only ever be "feed" in this case
func (st *State) ClientOther(filename string, client string) (*os.File, error) {

	if filename != "feed" {
		return nil, nil
	}
	current := st.clients[client]
	// We have the channel by name, now we need to make teh path.
	filePath := path.Join(*inPath, current.server, current.channel)
	return os.Open(filePath)
}

// ClientConnect - add last server in list, first channel in list
func (st *State) ClientConnect(client string) {
	def := st.clients["default"]
	st.clients[client] = &Client{server: def.server, channel: def.channel}
}

// ClientDisconnect - called when client disconnects
func (st *State) ClientDisconnect(client string) {
	delete(st.clients, client)
}
