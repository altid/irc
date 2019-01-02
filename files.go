package main

import (
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
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
	// Ensure our base path exists for each server
	serverPath := path.Join(*base, c.Addr)
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

func deleteChannel(channel, server, log string) {
	chandir := path.Join(server, channel)
	feed := path.Join(chandir, "feed")
	if runtime.GOOS == "plan9" {
		command := exec.Command("/bin/unmount", feed)
		command.Run()
	}
	os.RemoveAll(chandir)
}
