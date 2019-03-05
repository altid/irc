package main

import (
	"fmt"
	"os"
	"path"

	"bitbucket.org/mischief/libauth"
	"github.com/mischief/ndb"
	"github.com/ubqt-systems/fslib"
)

type config struct {
	addr   string
	port   string
	filter string
	ssl    string
	chans  string
	log    string
	nick   string
	user   string
	name   string
	pass   string
}

func newConfig() (*config, error) {
	confdir, err := fslib.UserConfDir()
	if err != nil {
		return nil, err
	}
	filePath := path.Join(confdir, "ubqt.cfg")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}
	conf, err := ndb.Open(filePath)
	if err != nil {
		return nil, err
	}
	for _, rec := range conf.Search("service", "irc") {
		// Verify we're not on a different IRC server from the one requested
		if conf.Search("service", "irc").Search("address") != *srv {
			continue
		}
		return readRecord(rec)
	}
	return nil, fmt.Errorf("Unable to find record for %s\n", *srv)
}

func readRecord(rec ndb.Record) (*config, error) {
	datadir, err := fslib.UserShareDir()
	if err != nil {
		datadir = "/tmp/ubqt"
	}
	conf := &config{
		port:   "6667",
		ssl:    "none",
		log:    path.Join(datadir, "irc"),
		filter: "none",
	}
	for _, tup := range rec {
		switch tup.Attr {
		case "address":
			conf.addr = tup.Val
		case "port":
			conf.port = tup.Val
		case "filter":
			conf.filter = tup.Val
		case "channels":
			conf.chans = tup.Val
		case "log":
			conf.log = tup.Val
		case "auth":
			conf.pass = tup.Val
		case "nick":
			conf.nick = tup.Val
		case "user":
			conf.user = tup.Val
		case "name":
			conf.name = tup.Val
		case "ssl":
			conf.ssl = tup.Val
		}
	}
	if conf.log == "" {
		conf.log = path.Join(datadir, "ircfs")
	}
	if len(conf.pass) > 5 && conf.pass[:5] == "pass=" {
		conf.pass = conf.pass[5:]
	}
	if conf.pass == "factotum" {
		userPwd, err := libauth.Getuserpasswd(
			"proto=pass service=irc server=%s user=%s",
			conf.addr,
			conf.user,
		)
		if err != nil {
			return nil, err
		}
		conf.pass = userPwd.Password
	}
	return conf, nil
}
