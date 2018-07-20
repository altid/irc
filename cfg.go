package main

import (
	"text/template"
	
	"github.com/vaughan0/go-ini"
)

// TODO: Move all parsing here
type Cfg struct { 
	chanFmt *template.Template
	selfFmt *template.Template
	ntfyFmt *template.Template
	servFmt *template.Template
	highFmt *template.Template
	actiFmt *template.Template
	modeFmt *template.Template
}

func ParseFormat(conf ini.File) *Cfg {
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
	return &Cfg{
    	chanFmt: template.Must(template.New("chan").Parse(chanFmt)),
    	ntfyFmt: template.Must(template.New("ntfy").Parse(ntfyFmt)),
    	servFmt: template.Must(template.New("serv").Parse(servFmt)),
    	selfFmt: template.Must(template.New("self").Parse(selfFmt)),
    	highFmt: template.Must(template.New("high").Parse(highFmt)),
    	actiFmt: template.Must(template.New("acti").Parse(actiFmt)),
    	modeFmt: template.Must(template.New("mode").Parse(modeFmt)),
	}
}

