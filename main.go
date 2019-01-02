package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	InputMsg MessageType = iota
	ChanMsg
	SelfMsg
	ServerMsg
	NotifyMsg
	TitleMsg
	StatusMsg
	HighMsg
	ActionMsg
	None
)
type MessageType int

var (
	base = flag.String("p", "/tmp/irc", "Path for filesystem (Default /tmp/irc)")
)

func init() {
	os.MkdirAll(*base, 0755)
}

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	// Try to clean up all we can on exit
	defer os.RemoveAll(*base)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGKILL, syscall.SIGINT)
	go func() {
		for sig := range c {
			switch sig {
			case syscall.SIGKILL, syscall.SIGINT:
			 	os.RemoveAll(*base)
				os.Exit(0)
			}
		}
	}()
	// config.go
	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// files.go
	err = CreateDirs(config)
	if err != nil {
		log.Fatal(err)
	}

	// server.go
	servers := GetServers(config)
	servers.Run()
}
