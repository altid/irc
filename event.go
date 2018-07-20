package main

// TODO: All events aggregated and dispatched here.
func (st *State) Run() {
	OutLoop()
	go st.CtlLoop("default")
	st.InLoop()
}
