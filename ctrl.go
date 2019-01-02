package main

import (
	"fmt"
)

type Ctrl struct {
	// basically we want the various channels we want to listen on here
}

func NewCtrl(s *Servers) *Ctrl {
	return &Ctrl{}
}


func (c *Ctrl) Listen() {
	var ctrl string
	fmt.Scanln(&ctrl)
	fmt.Print(ctrl)
}
