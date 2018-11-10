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
	dirpath := path.Join(*inPath, server)
	// Make sure path to file exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		os.MkdirAll(dirpath, 0755)
	}
	f, err := os.OpenFile(path.Join(dirpath, filename), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
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
