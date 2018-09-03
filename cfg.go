package main

import (
	"log"
	"net"
	"text/template"
	
	"github.com/vaughan0/go-ini"
	"github.com/go-irc/irc"
)

// Hold our default configurations
type Format struct { 
	chanFmt *template.Template
	selfFmt *template.Template
	ntfyFmt *template.Template
	servFmt *template.Template
	highFmt *template.Template
	actiFmt *template.Template
	modeFmt *template.Template
}

func GetFormat(conf ini.File) *Format {
    //Set some pretty printed defaults
    chanFmt := `[#5F87A7]({{.Name}}) {{.Data}}`
    selfFmt := `[#076678]({{.Name}}) {{.Data}}`
    highFmt := `[#9d0007]({{.Name}}) {{.Data}}`
    ntfyFmt := `[#5F87A7]({{.Name}}) {{.Data}}`
    servFmt := `--[#5F87A7]({{.Name}}) {{.Data}}--`
    actiFmt := `[#5F87A7( \* {{.Name}}) {{.Data}}`
    modeFmt := `--[#787878](Mode [{{.Data}}] by {{.Name}})`
    for key, value := range conf["options"] {
        switch key {
        case "channelfmt":
            chanFmt = value
        case "notificationfmt":
            ntfyFmt = value
        case "highfmt":
            highFmt = value
        case "selffmt":
            selfFmt = value
        case "actifmt":
            actiFmt = value
        case "modefmt":
            modeFmt = value
        }
    }
	return &Format{
    	chanFmt: template.Must(template.New("chan").Parse(chanFmt)),
    	ntfyFmt: template.Must(template.New("ntfy").Parse(ntfyFmt)),
    	servFmt: template.Must(template.New("serv").Parse(servFmt)),
    	selfFmt: template.Must(template.New("self").Parse(selfFmt)),
    	highFmt: template.Must(template.New("high").Parse(highFmt)),
    	actiFmt: template.Must(template.New("acti").Parse(actiFmt)),
    	modeFmt: template.Must(template.New("mode").Parse(modeFmt)),
	}
}

func GetConnection(conf ini.File, section string) (net.Conn, error) {
    server, ok := conf.Get(section, "Server")
    if ! ok {
        log.Println("Server entry not found!")
    }
    port, ok := conf.Get(section, "Port")
    if !ok {
        log.Println("No port set, using default")
        port = "6667"
    }
	addr := server + ":" + port
	return net.Dial("tcp", addr)
}

func GetConfig(conf ini.File, section string) (irc.ClientConfig, string) { 
    nick, ok := conf.Get(section, "Nick")
    if !ok {
        log.Println("nick entry not found")
    }
    user, ok := conf.Get(section, "User")
    if !ok {
        log.Println("user entry not found")
    }
    name, ok := conf.Get(section, "Name")
    if !ok {
        log.Println("name entry not found")
    }
    pw, ok := conf.Get(section, "Password")
    if !ok {
        log.Println("password entry not found")
    }
	buffers, ok := conf.Get(section, "Channels")
	if !ok {
		log.Println("Write `join #mychannel` to the ctl file to connect to channels")
	}
    return irc.ClientConfig{Nick: nick, User: user, Name: name, Pass: pw}, buffers
}
