package main

import (
	"log"
	"os"
)

func main() {
	test()
	log.Fatal("test")
	os.Exit(1)
	panic("test") // want "panic detected, usage of panic is forbidden"
}

func test() {
	log.Fatal("test") // want "log.Fatal detected outside of main function ; forbidden usage in function test outside of main function"
	os.Exit(1)        // want "os.Exit detected outside of main function ; forbidden usage in function test outside of main function"
	panic("test")     // want "panic detected, usage of panic is forbidden"
}
