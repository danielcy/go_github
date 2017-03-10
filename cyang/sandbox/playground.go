package main

import (
	"fmt"
	"sync"
)

var (
	recieverMutex *sync.Mutex = new(sync.Mutex)
	lockSignal    bool        = true
)

func Writer(m chan string) {
	<-m
	fmt.Printf("I'm writing... \n")
	fmt.Printf("Writing complete. \n")
	recieverMutex.Unlock()
}

func Reciever() {
	for {
		recieverMutex.Lock()
		fmt.Printf("I'm recieving... \n")
	}
}

func main() {

	m := make(chan string)

	go Reciever()
	go Writer(m)

	fmt.Printf("Ready to write. \n")
	m <- "a"
}
