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

type Keycmd struct {
    action string
    key string
    resp chan string
}
    

const KEYLEN = 10
const SERVERID = "walrus"

var Keyctrl chan Keycmd

func main() {
    flag.Parse()
    f, _ := os.OpenFile("walrus.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
    log.SetOutput(f)
    defer func() {
        f.Close()
    }()
    log.Println("-- Starting ", SERVERID, "--")
    Keyctrl = make(chan Keycmd)
    go randomkeygenerator(Keyctrl)

	// boilerplate code
	go control.start()
	http.HandleFunc("/", socketStart)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Println("server kon niet gestart worden het is niet gelukt.. helaas")
	}

	panic("just checking")
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
        todo = <- c
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
            case "short":
                for i := 3; i < len(todo.key); i++ {
                    count := 0
                    for key := range usedkeys {
                        if todo.key[0:i] == key[0:i] {
                            count++
                        }
                    }
                    if count == 1 {
                        todo.resp <- todo.key[0:i]
                        break
                    }
                }
        }
	}
}
