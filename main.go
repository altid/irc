package main

import (
	"flag"
	"log"
	"os"
	"path"
)

var (
	conf    = flag.String("c", "irc.ini", "Configuration file")
	inPath  = flag.String("p", path.Join(os.Getenv("HOME"), "irc"), "Path for file system")
	verbose = flag.Bool("v", false, "Enable verbose output")
)

// BUG: There's currently no checking in the setup phase of the ctl files. It will read and execute all commands it finds
// TODO: Scan to end of file on start before reading

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}
	st := newState()
	err := st.OutLoop()
	if err != nil {
		log.Fatal(err)
	}
	go st.CtlLoop("default")
	st.InLoop()
}
