package importalias

import (
	mylog "log"
	myos "os"
)

func panicCheck() {
	panic(nil) // want "panic call"

}

func exitCheck() {
	myos.Exit(1) // want "os.Exit call outside of main"
}

func fatalCheck() {
	mylog.Fatal() // want "log.Fatal call outside of main"
}
