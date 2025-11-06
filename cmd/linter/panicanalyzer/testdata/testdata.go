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

func DoPanic() {
	panic(nil)  // want "panic call"
	os.Exit(1)  // want "os.Exit call outside of main"
	log.Fatal() // want "log.Fatal call outside of main"
}
