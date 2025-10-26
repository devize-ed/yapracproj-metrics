package test

import (
	"log"
	"os"
)

func test() {
	log.Fatal("test") // want "log.Fatal detected outside of main function ; forbidden usage outside of main package"
	os.Exit(1)        // want "os.Exit detected outside of main function ; forbidden usage outside of main package"
	panic("test")     // want "panic detected, usage of panic is forbidden"
}
