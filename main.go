package main

import (
	"crypto/rand"
	"encoding/base32"
	"flag"
	"fmt"
	"net/http"
)

var sslport *int = flag.Int("sslport", 11012, "Port to listen.")
var sslcrt *string = flag.String("sslcert", "/etc/nginx/ssl/wss_erwin_io.pem", "ssl certificate")
var sslkey *string = flag.String("sslkey", "/etc/nginx/ssl/wss_erwin_io.key", "ssl private key")

type Keycmd struct {
	action string
	key    string
	resp   chan string
}

const KEYLEN = 10
const SERVERID = "walrus"

var Keyctrl chan Keycmd

func main() {
	flag.Parse()
	Keyctrl = make(chan Keycmd)
	go randomkeygenerator(Keyctrl)

	// boilerplate code
	go control.start()
	http.HandleFunc("/", socketStart)
	err := http.ListenAndServeTLS(fmt.Sprintf(":%d", *sslport), *sslcrt, *sslkey, nil)
	if err != nil {
		fmt.Println("server kon niet gestart worden het is niet gelukt.. helaas")
	}
}

func randomkeygenerator(c chan Keycmd) {
	var currentkey string
	var todo Keycmd
	b := make([]byte, KEYLEN)
	en := base32.StdEncoding
	usedkeys := make(map[string]bool)
	for {
	retry:
		rand.Read(b)
		d := make([]byte, en.EncodedLen(len(b)))
		en.Encode(d, b)
		currentkey = string(d)
		if usedkeys[currentkey] {
			goto retry
		}
		//action key response
		todo = <-c
		switch todo.action {
		case "get":
			todo.resp <- currentkey
			usedkeys[currentkey] = true
			close(todo.resp)
		case "del":
			todo.resp <- "ok"
			if usedkeys[todo.key] {
				usedkeys[todo.key] = false
				delete(usedkeys, todo.key)
			}
			close(todo.resp)
		case "add":
			todo.resp <- "ok"
			if !usedkeys[todo.key] {
				usedkeys[todo.key] = true
			}
		}
	}
}
