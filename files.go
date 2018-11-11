package main

import(
	"fmt"
	"log"
	"os"
	"path"
	"text/template"

)

type Data struct {
	Name string
	Message string
}

func WriteToFile(d *Data, server string, filename string, format *template.Template) {
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
	err = format.Execute(f, d)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(f, "\n")
}
