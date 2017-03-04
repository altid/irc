package main

import (
	"bytes"
	"fmt"
)

//TODO: Block on all reads for main files until any event occurs, then unlock -
// either all channels, or just the ones we're interested in.
// Reading title 200 times just because of a privmsg may be silly.
//TODO: So, wit hthat in mind, set up buffered channels for each file that needs to wait on relevent data. We may have events where titles are updated, or on buffer changes that unlock title, or people joining or parting that unlock sidebar, etc.
//TODO: Research how to best block reads here; do we need seperate goroutines or not?

// handleInput - append valid runes to input type, curtail input at [history]input lines.
func (st *State) handleInput(data []byte, client string) (int, error) {
	// Strip out initial forward slash of command, test for literal slash input
	if data[0] == '/' {
		data = data[1:]
		if data[0] != '/' {
			return st.handleCtl(data, client)
		}
	}
	current := st.clients[client]
	irc := st.irc[current.server]
	irc.Commands.Message(current.channel, string(data))
	st.input = append(st.input, data...)
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

func (st *State) ctl(client string) ([]byte, error) {
	return []byte("part\njoin\nquit\nbuffer\nignore\n"), nil
}

func (st *State) status(client string) ([]byte, error) {
	var buf []byte
	current := st.clients[client]
	irc := st.irc[current.server]
	channel := irc.Lookup(current.channel)
	if channel == nil {
		return nil, nil
	}
	//TODO: text/template to design the status bar
	buf = append(buf, '\\')
	buf = append(buf, []byte(channel.Name)...)
	buf = append(buf, []byte(channel.Modes.String())...)
	buf = append(buf, '\n')
	return buf, nil
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

func (st *State) buff(client string) ([]byte, error) {
	//TODO: Format either here, or have the logs formatted.
	//os.Open() make path based whichever current thing we're on
	//TODO: Update tabs to reflect
	return []byte("buffer file\n"), nil
}

func (st *State) title(client string) ([]byte, error) {
	current := st.clients[client]
	irc := st.irc[current.server]
	channel := irc.Lookup(current.channel)
	buf := []byte(channel.Topic)
	buf = append(buf, '\n')
	return buf, nil
}

func (st *State) handleIrc(server string, b []byte) {
	//TODO: All messages from IRC will filter through here, drawing to their respective files
	//TODO: Update tabs to reflect what occurs here
	fmt.Println(string(b))
}
