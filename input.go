package main

import (
	"fmt"
)

func inputHandler(st *State) {
	for {
		select {
		case input := <-st.event:
			fmt.Println(input)
		}
	}
}
