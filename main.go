package main

import (
	"flag"
	"log"
	"path"
	"os"

	"github.com/vaughan0/go-ini"
	"github.com/go-irc/irc"
)

var (
	config  = flag.String("c", "irc.ini", "Configuration file")
	inPath  = flag.String("p", path.Join(os.Getenv("HOME"), "irc"), "Path for file system")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	conf, err := ini.LoadFile(*config)
	if err != nil {
		log.Fatal(err)
	}

	// Main template stuff
	format := GetFormat(conf)

	// Parse each server entry
	for section := range conf {
		if section == "options" {
			continue
		}
		conn, err := GetConnection(conf, section)
		if err != nil {
			log.Printf("Error on server %s, %s\n", section, err)
		}
		serveraddr := GetSrvAddr(conf, section)
		config, buffers := GetConfig(conf, section)
		config.Handler = InitHandler(buffers, serveraddr, format)
		client := irc.NewClient(conn, config)
		// Start up input listeners here
		//InitInput(buffers, format, serveraddr)
		//go client.Run()
		client.Run()
	}

	// Start up control listener in final loop

}
