package main

import (
	"github.com/lrstanley/girc"
)

func (st *State) getChannel(client string) *girc.Channel {
	c := st.c[client]
	channel := c.irc.Lookup(c.channel)
	return channel
}

// handleInput - append valid runes to input type, curtail input at [history]input lines.
func (st *State) handleInput(data []byte, client string) (int, error) {
	// TODO: Scrub and send message to channel
	c := st.c[client]
	if c.channel != "" {
		c.irc.Commands.Message(c.channel, string(data))
	}
	st.input = append(st.input, data...)
	return len(data), nil
}

func (st *State) handleCtl(data []byte, name string) (int, error) {
	//TODO: Handle command
	switch data[0] {
	case 'j': //join
	case 'p': //part
	case 'q': //quit
	case 'b': //buffer
	}
	return len(data), nil
}

func (st *State) ctl(client string) ([]byte, error) {
	return []byte("part\njoin\nquit\nbuffer\n"), nil
}

func (st *State) status(client string) ([]byte, error) {
	var buf []byte
	channel := st.getChannel(client)
	if channel == nil {
		return nil, nil
	}
	buf = append(buf, '\\')
	buf = append(buf, []byte(channel.Name)...)
	buf = append(buf, []byte(channel.Modes.String())...)
	buf = append(buf, '\n')
	return buf, nil
}

func (st *State) sidebar(client string) ([]byte, error) {
	channel := st.getChannel(client)
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
	//os.Open() make path based whichever current thing we're on
	return []byte("buffer file\n"), nil
}

func (st *State) title(client string) ([]byte, error) {
	channel := st.getChannel(client)
	if channel == nil {
		return nil, nil
	}
	return []byte(channel.Topic), nil
}
