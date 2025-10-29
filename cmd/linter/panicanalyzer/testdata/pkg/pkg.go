package pkg

import (
	"log"
	"os"
)

func panicCheck() {
	panic(nil) // want "panic call"

}

func exitCheck() {
	os.Exit(1) // want "os.Exit call outside of main"
}

func fatalCheck() {
	log.Fatal() // want "log.Fatal call outside of main"
}
