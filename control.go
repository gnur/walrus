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
            Delkey <- c.clientid
			if len(control.clients[c.groupid]) == 0 {
				delete(control.clients, c.groupid)
                Delkey <- c.groupid
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
            if m.toid == SERVERID {
                if m.text == "getgroupid" {
                    control.clients[m.groupid][m.fromid].send <-SERVERID + ":" + m.groupid
                } else if m.text == "getclientid" {
                    control.clients[m.groupid][m.fromid].send <-SERVERID + ":" + m.fromid
                } else if m.text == "getallclientids" {
                    returntext := ""
                    for clientid, _ := range control.clients[m.groupid] {
                        returntext += clientid + ","
                    }
                    if returntext == "" {
                        returntext = ", "
                    }
                    fmt.Println(returntext)
                    control.clients[m.groupid][m.fromid].send <-SERVERID + ":" + returntext[:len(returntext)-1]
                }
            } else if m.toid != "" {
                if target, ok :=control.clients[m.groupid][m.toid]; ok {
                    target.send <-m.fromid + ":" + m.text
                }
            } else {
				for clientid, client := range control.clients[m.groupid] {
                    if clientid != m.fromid {
                        client.send <- m.fromid + ":" + m.text
                    }
				}
			}
		}
	}
}
