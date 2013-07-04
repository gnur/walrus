package main

import (
	"fmt"
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
			c.socket.SetReadDeadline(time.Now().Add(readWait))
		case websocket.OpText:
			incoming, err := ioutil.ReadAll(r)
			if err == nil {
				var msg = message{
					groupid: c.groupid,
					fromid:  c.clientid,
					text:    string(incoming),
				}
				fmt.Println(msg.text)
				control.msg <- msg
			} else {
				fmt.Println(err)
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

func socketStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		fmt.Println("GET")
		return
	}
	ws, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		fmt.Println("geen websocket")
		return
	} else if err != nil {
		log.Println(err)
		fmt.Println(err)
		return
	}
	parts := strings.Split(r.URL.Path[1:], "/")
	groupid := ""
	if match, er := regexp.MatchString("^[A-Z0-9]{16}$", parts[0]); er == nil && match {
		groupid = parts[0]
	}
	c := &connection{
		send:     make(chan string),
		clientid: <-Randkey,
		groupid:  groupid,
		socket:   ws,
	}
    log.Println(c.clientid, "connected from", r.RemoteAddr)
	control.connect <- c
	go c.writer()
	c.read()
}
