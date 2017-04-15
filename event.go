package main

import "github.com/ubqt-systems/ubqtlib"

func listenEvents(st *State, srv *ubqtlib.Srv) {
	for {
		select {
		case buf := <-st.event:
			go srv.SendEvent(buf)
		}
	}
}
