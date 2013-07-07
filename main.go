package main

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/http"
    "os"
    "log"
    "flag"
)

var port *int = flag.Int("p", 11011, "Port to listen.")

const KEYLEN = 10

var Randkey chan string
var Delkey chan string
var Closing chan chan bool
var Addkey chan string

func main() {
    flag.Parse()
    f, _ := os.OpenFile("walrus.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
    log.SetOutput(f)
    defer func() {
        f.Close()
    }()
	Randkey = make(chan string)
	Delkey = make(chan string)
	Closing = make(chan chan bool)
    Addkey = make(chan string)
	go randomkeygenerator(Randkey, Delkey, Closing, Addkey)

	// boilerplate code
	go control.start()
	http.HandleFunc("/", socketStart)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		fmt.Println("het is niet gelukt.. helaas")
	}

	closed := make(chan bool)
	Closing <- closed
	<-closed

	panic("just checking")
}

func randomkeygenerator(c chan string, del chan string, closing chan chan bool, add chan string) {
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
        case key := <-add:
            if !usedkeys[key] {
                usedkeys[key] = true
            }
		case t := <-closing:
			fmt.Println("killing myself now")
			t <- true
			return
		}
	}
}
