package main

import (
	"os"

	"github.com/coalaura/plain"
)

func main() {
	pl := plain.New(os.Stdout)

	pl.Println("Hello from Print")
	pl.Warnln("Hello from Warn")
	pl.Errorln("Hello from Error")

	input, _ := pl.Read(os.Stdin, "Input: ", 10)

	pl.Printf("You entered '%s'\n", input)

	options := []string{"Red", "Green", "Blue", "Yellow"}

	index, _ := pl.Select("Select: ", options)

	pl.Printf("You selected '%s'\n", options[index])
}
