package main

import (
	"log"
	"os"
)

func main() {
	panic(nil) // want "panic call"
	os.Exit(1)
	log.Fatal()
}
