package main

import(
	"fmt"
	"log"
	"os"
	"path"
	"text/template"

)

type Msg struct {
	Name string
	Data string
}

func WriteToFile(msg *Msg, server string, filename string, format *template.Template) {
	f, err := os.OpenFile(path.Join(*inPath, server, filename), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		log.Println(err)
		return
	}
	err = format.Execute(f, msg)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(f, "\n")
}
