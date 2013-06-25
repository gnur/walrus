package main

import (
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "time"
)

const KEYLEN = 10
func main() {
    randkey := make(chan string)
    delkey := make(chan string)
    closing := make(chan chan bool)
    go randomkeygenerator(randkey, delkey, closing)
    for i := 0; i < 5; i++ {
        key := <- randkey
        fmt.Println(key)
        delkey <- key
        time.Sleep(1 * time.Millisecond)
    }
    closed := make(chan bool)
    closing <- closed
    <- closed


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
        case key := <- del:
            if usedkeys[key] {
                usedkeys[key] = false
                delete(usedkeys, key)
            }
        case t := <-closing:
            fmt.Println("killing myself now")
            fmt.Println(usedkeys)
            t <- true
            return
        }
    }
}
