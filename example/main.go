package main

import (
	"os"
	"os/signal"

	"github.com/coalaura/plain"
)

func main() {
	pl := plain.New()

	// optional, prevents leftover un-reset colors
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		<-ch

		pl.Close()
		os.Exit(1)
	}()

	pl.Debugln("Hello from Debug")
	pl.Println("Hello from Print")
	pl.Warnln("Hello from Warn")
	pl.Errorln("Hello from Error")

	input, _ := pl.Read(os.Stdin, "Input: ", 64)

	pl.Printf("You entered '%s'\n", input)

	options := []string{"Red", "Green", "Blue", "Yellow"}

	index, _ := pl.Select("Select: ", options)

	pl.Printf("You selected '%s'\n", options[index])
}
