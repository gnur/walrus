package main

import "log"

type controlstruct struct {
	clients    map[string]map[string]*connection
	connect    chan *start
	disconnect chan *connection
	msg        chan message
}

type start struct {
    connection     *connection
    response    chan bool
}

var control = controlstruct{
	connect:    make(chan *start),
	disconnect: make(chan *connection),
	msg:        make(chan message),
	clients:    make(map[string]map[string]*connection),
}

func (control *controlstruct) start() {
	for {
		select {
		case c := <-control.disconnect:
			close(c.send)
			delete(control.clients[c.groupid], c.clientid)
            ch := make(chan string)
            Keyctrl <- Keycmd{action: "del", key:c.clientid, resp: ch}
            <-ch
			if len(control.clients[c.groupid]) == 0 {
				delete(control.clients, c.groupid)
                ch := make(chan string)
                Keyctrl <- Keycmd{action: "del", key:c.groupid, resp: ch}
                <-ch
				log.Println("all clients from", c.groupid, "disconnected")
			}
		case start := <-control.connect:
            c := start.connection
			if c.groupid == "" {
                ch := make(chan string)
                Keyctrl <- Keycmd{action: "get", resp: ch}
				c.groupid = <-ch
			}
			if _, ok := control.clients[c.groupid]; !ok {
				control.clients[c.groupid] = make(map[string]*connection)
			}
            if _, exists := control.clients[c.groupid][c.clientid]; exists {
                c.socket.Close()
                start.response <- false
            } else {
                control.clients[c.groupid][c.clientid] = c
                start.response <- true
            }
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
                    control.clients[m.groupid][m.fromid].send <-SERVERID + ":" + returntext[:len(returntext)-1]
                }
            } else if m.toid != "" {
                if target, ok :=control.clients[m.groupid][m.toid]; ok {
                    target.send <-m.fromid + ":" + m.text
                }
            } else {
				for _, client := range control.clients[m.groupid] {
                    client.send <- m.fromid + ":" + m.text
				}
			}
		}
	}
}
