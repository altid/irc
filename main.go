package main

import (
	"fmt"
	"flag"
	"os"
	"path"
)

var (
	conf    = flag.String("c", "irc.ini", "Configuration file")
	inPath  = flag.String("p", path.Join(os.Getenv("HOME"), "irc"), "Path for file system")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st := newState()
	err := st.OutLoop()
	if err != nil {
		// TODO: Use log
		fmt.Println(err)
		os.Exit(1)
	}
	st.InLoop()
}
