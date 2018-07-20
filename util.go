package main

import (
	"path"
	"os"
)

func FileCleanup(name string) error {
	filePath := path.Join(*inPath, name)
	return os.Remove(filePath)
}
