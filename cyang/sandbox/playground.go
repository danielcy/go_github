package main

import (
	"log"
)

func main() {
	s := "a"
	for i := 0; i < 10; i++ {
		if i == 5 {
			s = "b"
		}
		log.Printf(s)
	}
}
