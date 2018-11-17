package main

import(
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"text/template"
)

type Data struct {
	Name string
	Message string
	file string
	srv string
}

func NewData(name, message, srv, fileprefix, filesuffix string) *Data {
	filePath := path.Join(fileprefix, filesuffix)
	return &Data{Name: name, Message: message, file: filePath, srv: srv}
}

// Make sure we have good paths for file writes
func init() {
	if _, err := os.Stat(*inPath); os.IsNotExist(err) {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		*inPath = path.Join(usr.HomeDir, *inPath)
	}
	if _, err := os.Stat(*config); os.IsNotExist(err) {
	usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		*config = path.Join(usr.HomeDir, *config)
	}
}

func writeToFile(d *Data, format *template.Template) {
	filepath := path.Join(*inPath, d.srv, d.file)
	dirpath := path.Dir(filepath)
	// Make sure path to file exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		os.MkdirAll(dirpath, 0755)
	}
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
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

func msgToFile(buff, msg string) {
	filepath := path.Join(*inPath, buff)
	dirpath := path.Dir(filepath)
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		os.MkdirAll(dirpath, 0755)
	}
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(f, msg + "\n")
}

func setTopic(srv, buff, topic string) {
	filepath := path.Join(*inPath, srv, buff, "title")
	dirpath := path.Dir(filepath)
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		os.MkdirAll(dirpath, 0755)
	}
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(f, topic + "\n")
}
// TODO: Input loop here
