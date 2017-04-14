package main

import (
	"bytes"
	"os"
	"path"

	"github.com/ubqt-systems/cleanmark"
)

// handleInput - append valid runes to input type, curtail input at [history]input lines.
func (st *State) handleInput(data []byte, client string) (int, error) {
	// Strip out initial forward slash of command, test for literal slash input
	if data[0] == '/' {
		data = data[1:]
		if data[0] != '/' {
			return st.handleCtl(data, client)
		}
	}
	// We have a regular channel message at this point - make sure to remove tabs entry
	current := st.clients[client]
	if _, ok := st.tablist[current.channel]; ok {
		st.Lock()
		delete(st.tablist, current.channel)
		st.Unlock()
	}
	// Find our server, send formatted message
	irc := st.irc[current.server]
	irc.Commands.Message(current.channel, string(data))
	st.input = append(st.input, data...)
	// Ensure we log our data
	filePath := path.Join(*inPath, current.server, current.channel)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	if err != nil {
		return 0, err
	}
	e := &message{Data: string(cleanmark.CleanBytes(data)), Name: irc.Config.Nick}
	err = st.selfFmt.Execute(f, e)
	if err != nil {
		return 0, err
	}
	// Finally notify our events file of the update
	st.event <- []byte("main\n")
	return len(data), nil
}

func (st *State) handleCtl(b []byte, client string) (int, error) {
	arr := bytes.Fields(b)
	switch string(arr[0]) {
	case "set":
		// Set for client specif
		st.handleSet(arr[1:], client)
	// Handle -server, default to current [client]
	case "q":
		message := bytes.Join(arr[2:], []byte(" "))
		st.handleMsg(string(arr[1]), string(message), client)
	case "msg":
		message := bytes.Join(arr[2:], []byte(" "))
		st.handleMsg(string(arr[1]), string(message), client)
	case "join":
		// We only need current irc connection here
		st.handleJoin(string(arr[1]), client)
	case "part":
		// We only need current irc connection here
		st.handlePart(string(arr[1]), client)
	case "buffer":
		// Buffer swapping
		st.handleBuffer(string(arr[1]), client)
	case "ignore":
		// This will be a global blacklist that we just don't log messages with, won't need client. Will just be `st.AddIgnore(b) and such
		// Store to file, such as `irc/freenode/ignore`
		st.handleIgnore(arr[1:], client)
	case "connect":

	}
	return len(b), nil
}

func (st *State) status(client string) ([]byte, error) {
	current := st.clients[client]
	irc := st.irc[current.server]
	channel := irc.Lookup(current.channel)
	if channel == nil {
		return nil, nil
	}
	//TODO: text/template to design the status bar
	stat := "[#587624]( mode )" + channel.Modes.String() + "\n"
	return []byte(stat), nil
}

func (st *State) sidebar(client string) ([]byte, error) {
	current := st.clients[client]
	irc := st.irc[current.server]
	channel := irc.Lookup(current.channel)
	if channel == nil {
		return nil, nil
	}
	var buf []byte
	list := channel.NickList()
	for _, item := range list {
		buf = append(buf, []byte(item)...)
		buf = append(buf, '\n')
	}
	return buf, nil
}

func (st *State) title(client string) ([]byte, error) {
	current := st.clients[client]
	irc := st.irc[current.server]
	channel := irc.Lookup(current.channel)
	buf := []byte(channel.Topic)
	buf = append(buf, '\n')
	return buf, nil
}

func (st *State) tabs(client string) ([]byte, error) {
	current := st.clients[client]
	buf := "[#587624](" + current.channel + ")"
	for item, color := range st.tablist {
		if item != current.channel {
			buf += " " + color + "(" + item + ")"
		}
	}
	buf += "\n"
	return []byte(buf), nil
}
