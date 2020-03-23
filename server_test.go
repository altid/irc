package main

import (
	"context"
	"testing"

	"github.com/altid/libs/fs"
	"github.com/altid/libs/markup"
)

func TestCmds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &server{
		cancel: cancel,
		i:      make(chan string),
		e:      make(chan string),
		j:      make(chan string),
		m:      make(chan *msg),
		debug:  func(ctlItem, ...interface{}) {},
	}

	reqs := make(chan string)

	mcf, err := fs.MockCtlFile(ctx, s, reqs, "test", true)
	if err != nil {
		t.Error(err)
		return
	}

	defer mcf.Cleanup()

	go s.fileListener(ctx, mcf)
	go runCommands(reqs)
	go mcf.Listen()

	<-ctx.Done()
}

func runCommands(reqs chan string) {
	reqs <- "open foo"
	reqs <- "open bar"
	// Stuff like me will have to have major parts abstracted out
	// Since it calls a conn, we'll have to swap it for
	// a readwriter, etc
	//reqs <- "join baz"
	//reqs <- "join qoz"
	//reqs <- "part baz"
	//reqs <- "me foo smiles"
	reqs <- "test quit"
}

type mockhandler struct{}

func (f *mockhandler) Handle(bufname string, l *markup.Lexer) error {
	_, err := input(l)
	if err != nil {
		return err
	}

	return nil
}

func TestServerInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &server{
		cancel: cancel,
		i:      make(chan string),
		e:      make(chan string),
		j:      make(chan string),
		m:      make(chan *msg),
		debug:  func(ctlItem, ...interface{}) {},
	}

	reqs := make(chan string)

	mcf, err := fs.MockCtlFile(ctx, s, reqs, "test", true)
	if err != nil {
		t.Error(err)
		return
	}

	defer mcf.Cleanup()
	go s.fileListener(ctx, mcf)

	go func() {
		mcf.CreateBuffer("foo", "feed")
		in := make(chan string)

		input, err := fs.NewMockInput(&mockhandler{}, "foo", false, in)
		if err != nil {
			t.Error(err)
		}

		input.Start()
		defer input.Stop()
		in <- "test some"
		in <- "input"
		in <- "make some"
		in <- "things break"
		//https://github.com/altid/libs/issues/13
		//in <- "invalid-tokens"

		if e := input.Errs(); len(e) > 0 {
			for _, err := range e {
				t.Error(err)
			}
		}
		reqs <- "test quit"
	}()

	go mcf.Listen()

	<-ctx.Done()
}
