package main

import (
	"io"

	"github.com/altid/libs/config"
)

func buildConfig(rw io.ReadWriter) (*config.Config, error) {
	repl := struct {
		Address string `IP Address of service`
		Port    int    `Port to use`
		Ssl     bool   `Do you wish to use SSL?`
		Auth    string `Auth to use: pass=mypass|factotum`
		Nick    string `Nickname`
		User    string `User name`
		Name    string `Real name`
	}{"irc.freenode.net", 6667, false, "pass=hunter2", "Guest", "Guest", "Guest"}

	return config.Repl(rw, repl, false)
}
