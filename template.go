package main

import (
	"text/template"

	"github.com/mischief/ndb"
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
	titlFmt *template.Template
}

func getTemplate(format string, name string) *template.Template {
	return template.Must(template.New(name).Parse(format))
}

func GetFormats(ndb *ndb.Ndb) *Format {
	//Set some pretty printed defaults
	chanFmt := `[lightblue]({{.Name}}) {{.Message}}`
	selfFmt := `[blue]({{.Name}}) {{.Message}}`
	highFmt := `[red]({{.Name}}) {{.Message}}`
	ntfyFmt := `[lightblue]({{.Name}}) {{.Message}}`
	servFmt := `--[grey]({{.Name}}) {{.Message}}--`
	actiFmt := `[lightblue]( \* {{.Name}}) {{.Message}}`
	modeFmt := `--[grey](Mode [{{.Message}}] by {{.Name}})`
	titlFmt := `{{.Message}}`

	records := ndb.Search("format", "default")
	for _, rec := range records {
		for _, tuple := range rec {
			switch tuple.Attr {
			case "channel":
				chanFmt = tuple.Val
			case "notify":
				ntfyFmt = tuple.Val
			case "server":
				servFmt = tuple.Val
			case "self":
				selfFmt = tuple.Val
			case "highlight":
				highFmt = tuple.Val
			case "action":
				actiFmt = tuple.Val
			case "mode":
				modeFmt = tuple.Val
			case "title":
				titlFmt = tuple.Val
			}
		}
	}

	return &Format{
		chanFmt: getTemplate(chanFmt, "channel"),
		ntfyFmt: getTemplate(ntfyFmt, "notify"),
		servFmt: getTemplate(servFmt, "server"),
		selfFmt: getTemplate(selfFmt, "self"),
		highFmt: getTemplate(highFmt, "highlight"),
		actiFmt: getTemplate(actiFmt, "action"),
		modeFmt: getTemplate(modeFmt, "mode"),
		titlFmt: getTemplate(titlFmt, "title"),
	}
}
