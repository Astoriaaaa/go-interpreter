package main

import (
	"fmt"
	"interpreter/repl"
	"os"
)

func main() {
	fmt.Printf("This is Monkey Language\n")
	fmt.Printf("Type any commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
