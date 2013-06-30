package main

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/http"
)

const KEYLEN = 10

var Randkey chan string
var Delkey chan string
var Closing chan chan bool

func main() {
	Randkey = make(chan string)
	Delkey = make(chan string)
	Closing = make(chan chan bool)
	go randomkeygenerator(Randkey, Delkey, Closing)

	// boilerplate code
	go control.start()
	http.HandleFunc("/", socketStart)
	err := http.ListenAndServe(":11011", nil)
	if err != nil {
		fmt.Println("het is niet gelukt.. helaas")
	}

	closed := make(chan bool)
	Closing <- closed
	<-closed

	panic("just checking")
}

func randomkeygenerator(c chan string, del chan string, closing chan chan bool) {
	var currentkey string
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

		select {
		case c <- currentkey:
			usedkeys[currentkey] = true
		case key := <-del:
			if usedkeys[key] {
				usedkeys[key] = false
				delete(usedkeys, key)
			}
		case t := <-closing:
			fmt.Println("killing myself now")
			t <- true
			return
		}
	}
}
