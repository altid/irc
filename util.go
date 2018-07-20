package main

import (
	"path"
	"os"
)

func FileCleanup(name string) {
	filePath := path.Join(*inPath, name)
	err := os.Remove(filePath)
	return err
}
