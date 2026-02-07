package main

import (
	"github.com/coalaura/plain"
)

var pl = plain.New()

func main() {
	pl.Debugln("Hello from Debug")
	pl.Println("Hello from Print")
	pl.Warnln("Hello from Warn")
	pl.Errorln("Hello from Error")

	confirmed, err := pl.ConfirmWithEcho("Confirm", true, " ")
	pl.MustExit(err)

	if confirmed {
		pl.Println("You confirmed")
	} else {
		pl.Println("You declined")
	}

	input, err := pl.Read("Input: ", 64)
	pl.MustExit(err)

	pl.Printf("You entered '%s'\n", input)

	options := []string{"Red", "Green", "Blue", "Yellow"}

	index, err := pl.Select("Select: ", options)
	pl.MustExit(err)

	pl.Printf("You selected '%s'\n", options[index])
}
