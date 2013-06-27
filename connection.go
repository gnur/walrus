package main

import (
    "github.com/garyburd/go-websocket/websocket"
    "time"
    "io/ioutil"
)

const (
    readWait = 60 * time.Second
    writeWait = 10 * time.Second
    pingPeriod = (readWait * 9) / 10
)

type message struct {
    fromid string
    groupid string
    text string
}

type connection struct {
    groupid   string
    clientid  string
    socket    *websocket.Conn
    send      chan string
    err       error
}

func (c *connection) read() {
    defer func() {
        control.disconnect <- c
        c.socket.Close()
    }()
    c.socket.SetReadDeadline(time.Now().Add(readWait))
    for {
        op, r, err := c.socket.NextReader()
        if err != nil {
            c.err = err
            break
        }
        switch op {
            case websocket.OpPong:
                c.socket.SetReadDeadline(time.Now().Add(readWait))
            case websocket.OpText:
                incoming, err := ioutil.ReadAll(r)
                if err != nil {
                    var msg = message{
                        groupid: c.groupid,
                        fromid: c.clientid,
                        text: string(incoming),
                       } 
                    control.msg <- msg
                }
        }
    }
}
func (c *connection) write(opCode int, payload []byte) error {
        c.socket.SetWriteDeadline(time.Now().Add(writeWait))
        return c.socket.WriteMessage(opCode, payload)
}

func (c *connection) writer() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.socket.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.write(websocket.OpClose, []byte{})
                return
            }
            if err := c.write(websocket.OpText, []byte(message)); err != nil {
                c.err = err
                return
            }
        case <-ticker.C:
            if err := c.write(websocket.OpPing, []byte{}); err != nil {
                c.err = err
                return
            }
        }
    }
}

func socketStart(ws *websocket.Conn) {
    c := &connection{
            send: make(chan string),
            clientid: <-randkey,
            groupid: "",
            socket: ws,
    }
    go c.writer()
    c.read()
}
