package main

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"os"
	"path"
	"runtime"
	"text/template"

	"bitbucket.org/mischief/libauth"
	"github.com/mischief/ndb"
)

// ServerConf is parsed on an initial IRC connection, it holds no runtime state
type Config struct {
	Addr   string
	Port   string
	Filter string
	Ssl    string
	Chans  string
	Log    string
	Nick   string
	User   string
	Name   string
	Pass   string
	Theme  string
	Fmt    map[string]*template.Template
}

// GetConfig - return a usable *Config array
func NewConfig() ([]*Config, error) {
	confdir, err := userConfDir()
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
	servers, err := newServerConf(conf)
	if len(servers) == 0 {
		return nil, errors.New("No servers configured, bailing. Check log for details")
	}
	for _, server := range servers {
		server.Fmt = newFormat(server.Theme, conf)
	}
	return servers, err
}

// TODO: Switch on conf.Ssl, "none" or "", "simple", "/path/to/cert" or whatever is idiomatic
func GetConn(conf *Config) (net.Conn, error) {
	if conf.Ssl == "true" {
		tlsConfig := &tls.Config{
			ServerName:         conf.Addr + ":" + conf.Port,
			InsecureSkipVerify: true,
		}
		conn, err := tls.Dial("tcp", conf.Addr+":"+conf.Port, tlsConfig)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	conn, err := net.Dial("tcp", conf.Addr+":"+conf.Port)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Attempt to parse named format, will fall back to known working default template
func newFormat(theme string, conf *ndb.Ndb) map[string]*template.Template {
	records := conf.Search("format", theme)
	if theme == "default" {
		return defaultFormat()
	}
	if records == nil {
		log.Printf("theme not found: %s, using default", theme)
		return defaultFormat()
	}
	items := []string{
		"channel",
		"notify",
		"server",
		"highlight",
		"self",
		"action",
		"mode",
		"title",
	}
	format := make(map[string]*template.Template)
	for _, item := range items {
		record := records.Search(item)
		template, err := ParseFormat(item, record)
		if err != nil {
			log.Printf("unable to parse config: %s\n", err)
			log.Print("using default format")
			return defaultFormat()
		}
		format[item] = template
	}
	return format
}

func defaultFormat() map[string]*template.Template {
	format := make(map[string]*template.Template)
	format["highlight"], _ = ParseFormat("highlight", `{{.Time}} <[#9d0007]({{.Name}})> {{.Data}}`)
	format["channel"], _ = ParseFormat("channel", `{{.Time}} <[#5F87A7]({{.Name}})> {{.Data}}`)
	format["notify"], _ = ParseFormat("notify", `{{.Time}} <[#5F87A7]({{.Name}})> {{.Data}}`)
	format["server"], _ = ParseFormat("server", `{{.Time}} --[#5F87A7]({{.Name}}) {{.Data}}`)
	format["action"], _ = ParseFormat("action", `{{.Time}} [#5F87A7]( \* {{.Name}}) {{.Data}}`)
	format["title"], _ = ParseFormat("title", `[#5F87A7]({{.Data}})`)
	format["self"], _ = ParseFormat("self", `{{.Time}} <[#076678]({{.Name}})> {{.Data}}`)
	format["mode"], _ = ParseFormat("mode", `{{.Time}} <[#5F87A7]({{.Name}})> {{.Data}}`)
	return format
}

// ParseFormat - helper function to set Config.Fmt members
func ParseFormat(name, format string) (*template.Template, error) {
	// TODO: Error checking
	return template.Must(template.New(name).Parse(format)), nil
}

func newServerConf(conf *ndb.Ndb) ([]*Config, error) {
	datadir, err := userShareDir()
	if err != nil {
		datadir = "/tmp/ubqt"
	}
	var serverConfigs []*Config
	for _, rec := range conf.Search("service", "irc") {
		conf := &Config{
			Port:   "6667",
			Ssl:    "none",
			Log:    path.Join(datadir, "irc"),
			Filter: "none",
			Theme:  "default",
		}
		for _, tup := range rec {
			switch tup.Attr {
			case "address":
				conf.Addr = tup.Val
			case "port":
				conf.Port = tup.Val
			case "filter":
				conf.Filter = tup.Val
			case "channels":
				conf.Chans = tup.Val
			case "log":
				conf.Log = tup.Val
			case "auth":
				conf.Pass = tup.Val
			case "nick":
				conf.Nick = tup.Val
			case "user":
				conf.User = tup.Val
			case "name":
				conf.Name = tup.Val
			case "ssl":
				conf.Ssl = tup.Val
			}
		}
		if conf.Addr == "" {
			log.Print("Missing \"address=\" entry, unable to add server")
			continue
		}
		if conf.Log == "" {
			datadir, err := userShareDir()
			if err != nil {
				log.Printf("Unable to set up logs: %s\n", err)
				continue
			}
			conf.Log = path.Join(datadir, "ircfs")
		}
		if len(conf.Pass) > 5 && conf.Pass[:5] == "pass=" {
			conf.Pass = conf.Pass[5:]
		}
		if conf.Pass == "factotum" {
			UserPwd, err := libauth.Getuserpasswd("proto=pass service=irc server=%s user=%s", conf.Addr, conf.User)
			if err != nil {
				log.Print("Factotum not running and/or password entry not found - unable to connect")
				continue
			}
			conf.Pass = UserPwd.Password
		}
		serverConfigs = append(serverConfigs, conf)
	}
	return serverConfigs, nil
}

// Mimick UserCacheDir() with default Data/Config dirs
func userShareDir() (string, error) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}
	case "darwin":
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}
		dir += "/Library"
	case "plan9":
		dir = os.Getenv("home")
		if dir == "" {
			return "", errors.New("$home is not defined")
		}
		dir += "/lib"
	default: // Unix
		dir = os.Getenv("XDG_DATA_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				return "", errors.New("neither $XDG_DATA_HOME nor $HOME is defined")
			}
			dir += "/.local/share"
		}
	}
	return dir, nil
}

func userConfDir() (string, error) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}
	case "darwin":
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}
		dir += "/Library/Preferences"
	case "plan9":
		dir = os.Getenv("home")
		if dir == "" {
			return "", errors.New("$home is not defined")
		}
		dir += "/lib"
	default: // Unix
		dir = os.Getenv("XDG_CONFIG_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				return "", errors.New("neither $XDG_CONFIG_HOME nor $HOME is defined")
			}
			dir += "/.config"
		}
	}
	return dir, nil
}
