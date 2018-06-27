package main

// TODO: Change to a ctl loop per server (irc.freenode.net), leaving the client to granularily connect to whichever.
// Instead of inpath/irc, we will inpath/irc.freenode.net
// inpath/irc.freenode.net/ctl, event, etc instead of an aggregated one. We can still connect to various services through ircfs to keep configuration sane.
// This requires slight logic changes, for event handling and how ctl is set up - but aside from that will be simple to implement, and allow further generalization to the client.


import (
	"bufio"
	"bytes"
	"log"
	"path"
	"os"
	"time"
	"fmt"
)

func isValidRequest(b []byte, s string) bool {
	return bytes.HasPrefix(bytes.ToLower(b), []byte(s))

}

// TODO: Make sure we have enough context to act on correct server
func (st *State) Control(b []byte, server string) {
	// TODO: Check for irc.freenode.net/#ubqt vs #ubqt in girc
	switch b[0] {
	case 'j', 'J':
		if server == "default" {
			fmt.Println("here")
			break
		}
		fmt.Println("We here now")
		if isValidRequest(b, "join") {
			// TODO: Validate legal channel name, and that we don't have a password or multiple channels here. Seperate by word, test each word, then pass through after?
			// TODO: Break out joinin ga channel to a dedicated function for both initialization and future joins like here.
			fmt.Printf("Joining %s - %s\n", server, string(b[5:]))
			srv := st.irc[server]
			srv.Cmd.Join(string(b[5:]))
			filePath := path.Join(*inPath, server, (string(b[5:])))
			err := os.MkdirAll(filePath, 0777)
			if err != nil {
				log.Print(err)
			}
		}
	case 'p', 'P':
		if server == "default" {
			break
		}
		if isValidRequest(b, "part") {
			srv := st.irc[server]
			srv.Cmd.Part(string(b[5:]))
			// TODO: Validate that channel names are legal prior to sending request
		}
	case 'q', 'Q':
		if isValidRequest(b, "quit") {
			st.Cleanup()
			// Do a goto Cleanup, exit cleanly
		}
	case 'r', 'R':
		if isValidRequest(b, "reconnect") {
			// TODO: See if we have a config for the given server
			// Connect on success, or error "No such server configured"
		}
	case 'c', 'C':
		if isValidRequest(b, "connect") {
			// TODO: See if we have config for given server
			// Connect on success, or error "No such server configured"
		}
	}
}

// This is the main control loop that will listen for writes to
// *inpath/ctl, and act on those. Other listeners will exist on the ctl for each given server that we connect to.
func (st *State) CtlLoop(srv string) {
	filePath := path.Join(*inPath, srv, "ctl")
	if srv == "default" {
		filePath = path.Join(*inPath, "ctl")
	}
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	buffer := bufio.NewReader(f)
	// Cheapo epoll
	for {
		b, _, _ := buffer.ReadLine()
		if len(b) != 0 {
			st.Control(b, srv)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (st *State) Cleanup() {
	filePath := path.Join(*inPath, "ctl")
	err := os.Remove(filePath)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Send close to all channels left open
	// So we can remove ctl from all connected channels, etc.
}
