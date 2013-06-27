package main

type controlstruct struct {
	clients    map[string]map[string]*connection
	connect    chan *connection
	disconnect chan *connection
    msg        chan message
}

var control = controlstruct{
    connect:     make(chan *connection),
    disconnect:  make(chan *connection),
    msg:         make(chan message),
    clients:     make(map[string]map[string]*connection),
}

func (control *controlstruct) start() {
    for {
        select {
        case c := <- control.disconnect:
            close(c.send)
            delete(control.clients[c.groupid], c.clientid)
        case c := <- control.connect:
            if c.groupid == "" {
                c.groupid = <-randkey
            }
            control.clients[c.groupid][c.clientid] = c
        case m := <- control.msg:
            if m.groupid != "" {
                if m.fromid != "" {
                    control.clients[m.groupid][m.fromid].send <- m.text
                } else {
                    for _, client := range control.clients[m.groupid] {
                        client.send <- m.text
                    }
                }
            }
        }
    }
}
