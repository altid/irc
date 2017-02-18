package main

func getFirstWord(b []byte) string {
	for n, i := range b {
		if n == ' ' || n == '\n' || n == '\r' {
			return string(b[i:])
		}
	}
	return string(b)
}

func (st *State) handleSet(b []byte, client string) {

}

func (st *State) handleMsg(b []byte, client string) {

}

func (st *State) handleJoin(b []byte, client string) {

}

func (st *State) handlePart(b []byte, client string) {

}

func (st *State) handleBuffer(b []byte, client string) {

}

func (st *State) handleIgnore(b []byte, client string) {

}
