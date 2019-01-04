package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-irc/irc"
)

// Describes a valid IRC channel
var Match = regexp.MustCompile("([&#][^\\s\\x2C\\x07]{1,199})")

// For each channel on each server, initialize directory and logs
func CreateDirs(c []*Config) error {
	for _, config := range c {
		err := createDir(config)
		if err != nil {
			return err
		}
	}
	return nil
}

func createDir(c *Config) error {
	// Ensure our path exists for each server
	serverPath := path.Join(*mtpt, c.Addr)
	os.MkdirAll(path.Join(serverPath, "server"), 0755)

	// Ensure our log path exists for each server
	logPath := path.Join(c.Log, c.Addr)
	os.MkdirAll(logPath, 0755)

	channels := Match.FindAllString(c.Chans, -1)
	for _, channel := range channels {
		CreateChannel(channel, serverPath, logPath)
	}

	// And we create a file in our temporary dir for the server
	os.Mkdir(path.Join(serverPath, "server"), 0755)
	Touch(path.Join(serverPath, "server", "feed"))
	return nil
}

// Used externally as well for DMs in server.go
func CreateChannel(channel, server, log string) error {
	os.Mkdir(path.Join(server, channel), 0755)
	chanlog := path.Join(log, channel)
	Touch(chanlog)
	feed := path.Join(server, channel, "feed")
	// We don't log server messages
	if channel == "server" {
		Touch(feed)
		return nil
	}
	switch runtime.GOOS {
	case "plan9":
		Touch(feed)
		command := exec.Command("/bin/bind", chanlog, feed)
		err := command.Run()
		if err != nil {
			return err
		}
	default:
		err := os.Symlink(chanlog, feed)
		if err != nil {
			return err
		}
	}
	return nil
}

func Touch(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fp, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, 0644)
		fp.Close()
	}
}

func DeleteChannel(filename string) {
	if runtime.GOOS == "plan9" {
		command := exec.Command("/bin/unmount", filename)
		command.Run()
	}
	os.RemoveAll(path.Dir(filename))
}

func Event(filename string, s *Server) {
	file := path.Join(*mtpt, s.addr, "event")
	if _, err := os.Stat(path.Dir(file)); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(file), 0755)
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		log.Print(err)
		return
	}
	f.WriteString(filename + "\n")
}

func Title(filename string, server *Server, m *irc.Message) {
	file := path.Join(*mtpt, server.addr, filename)
	if _, err := os.Stat(path.Dir(file)); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(file), 0755)
	}
	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		log.Print(err)
		return
	}
	f.WriteString(m.Trailing() + "\n")
}

func WriteTo(fileName, whom string, server *Server, m *irc.Message, msgType MessageType) {
	srvdir := path.Join(*mtpt, server.addr)
	if _, err := os.Stat(path.Join(srvdir, fileName)); os.IsNotExist(err) {
		logdir := path.Join(server.log, server.addr)
		CreateChannel(path.Dir(fileName), srvdir, logdir)
		Touch(path.Join(srvdir, fileName))
	}
	format := parseForFormat(server, msgType)
	message, err := NewMessage(format, server, fileName, whom)
	defer message.Close()
	if err != nil {
		log.Printf("Invalid message %s %s\n", err, m.String())
		return
	}
	content := strings.NewReader(m.Trailing())
	_, err = content.WriteTo(message)
	Event(path.Join(srvdir, fileName), server)
}
