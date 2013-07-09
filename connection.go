package main

import (
	"github.com/garyburd/go-websocket/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	readWait   = 60 * time.Second
	writeWait  = 10 * time.Second
	pingPeriod = (readWait * 9) / 10
)

type message struct {
    toid    string
	fromid  string
	groupid string
	text    string
}

type connection struct {
	groupid  string
	clientid string
	socket   *websocket.Conn
	send     chan string
	err      error
}

func (c *connection) read() {
	defer func() {
		control.disconnect <- c
		c.socket.Close()
        log.Println(c.clientid, "disconnected")
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
            log.Println(c.clientid, "pong")
			c.socket.SetReadDeadline(time.Now().Add(readWait))
		case websocket.OpText:
			incoming, err := ioutil.ReadAll(r)
			if err == nil {
                parts := strings.SplitN(string(incoming), ":", 2)
                if parts != nil && len(parts) == 2 {
                    var msg = message{
                        toid: parts[0],
                        groupid: c.groupid,
                        fromid:  c.clientid,
                        text:    parts[1],
                    }
                    control.msg <- msg
                }
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
            log.Println(c.clientid, "ping")
			if err := c.write(websocket.OpPing, []byte{}); err != nil {
				c.err = err
				return
			}
		}
	}
}

func socketStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	parts := strings.Split(r.URL.Path[1:], "/")
	groupid := ""
    clientid := ""
	if match, er := regexp.MatchString("^[A-Z0-9]{16}$", parts[0]); er == nil && match {
		groupid = parts[0]
        Addkey <-groupid
	}
    if len(parts) > 1 {
        if match, er := regexp.MatchString("^[A-Z0-9]{16}$", parts[1]); er == nil && match {
            clientid = parts[1]
            Addkey <-clientid
        }
	}
    if clientid == "" {
        clientid = <-Randkey
    }
	c := &connection{
		send:     make(chan string),
		clientid: clientid,
		groupid:  groupid,
		socket:   ws,
	}
    log.Println(c.clientid, "connected from", r.RemoteAddr)
    response := make(chan bool)
    start := &start{
        connection: c,
        response: response,
    }
	control.connect <- start
    if <-response {
        close(response)
        go c.writer()
        c.read()
    } else {
        log.Println(c.clientid, "disconnected because id is already in use")
        close(response)
    }
}
