package main

import (
	"log"
	"os"
	"path"
	"sync"
	"text/template"
	"time"
)

func NewMessage(temp *template.Template, srv *Server, channel, name string) (*Message, error) {
	data := make(chan []byte)
	message := &Message{data: data, temp: temp}
	fp, err := message.open(srv, channel)
	if err != nil {
		return nil, err
	}
	go message.wait(fp, name)
	return message, nil
}

type Message struct {
	temp *template.Template
	data chan []byte
	sync.Mutex
}

func (m *Message) Write(p []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	b := make([]byte, len(p))
	n = copy(b, p)
	m.data <- b
	return n, nil
}

func (m *Message) Close() error {
	close(m.data)
	return nil
}

func (m *Message) wait(fp *os.File, name string) {
	defer fp.Close()
	for b := range m.data {
		body := string(b)
		time := time.Now().Format(time.RFC3339)
		msg := struct {
			Name string
			Data string
			Time string
		}{
			name,
			body + "\n",
			time,
		}
		err := m.temp.Execute(fp, msg)
		if err != nil {
			log.Printf("Error writing message: %s\n", err)
		}
	}
}

func (m *Message) open(srv *Server, name string) (*os.File, error) {
	filename := path.Join(*base, srv.addr, name)
	return os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
}
