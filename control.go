package main

import "fmt"

type controlstruct struct {
	clients    map[string]map[string]*connection
	connect    chan *connection
	disconnect chan *connection
	msg        chan message
}

var control = controlstruct{
	connect:    make(chan *connection),
	disconnect: make(chan *connection),
	msg:        make(chan message),
	clients:    make(map[string]map[string]*connection),
}

func (control *controlstruct) start() {
	for {
		select {
		case c := <-control.disconnect:
			fmt.Println("discconect")
			close(c.send)
			delete(control.clients[c.groupid], c.clientid)
			if len(control.clients[c.groupid]) == 0 {
				delete(control.clients, c.groupid)
				fmt.Println("all clients from", c.groupid, "disconnected")
			}
		case c := <-control.connect:
			if c.groupid == "" {
				c.groupid = <-Randkey
			}
			fmt.Println(c.groupid)
			if _, ok := control.clients[c.groupid]; !ok {
				control.clients[c.groupid] = make(map[string]*connection)
			}
			control.clients[c.groupid][c.clientid] = c
		case m := <-control.msg:
			if m.groupid != "" || m.groupid == "" {
				for _, client := range control.clients[m.groupid] {
					client.send <- m.text
				}
			}
		}
	}
}
