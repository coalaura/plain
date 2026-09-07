package main

import (
	"github.com/coalaura/plain"
)

type colorOption struct {
	label       string
	description string
}

func (o colorOption) Label() string {
	return o.label
}

func (o colorOption) Description() string {
	return o.description
}

func main() {
	pl := plain.New()

	pl.Debugln("Hello from Debug")
	pl.Println("Hello from Print")
	pl.Warnln("Hello from Warn")
	pl.Errorln("Hello from Error")

	confirmed, err := pl.ConfirmWithEcho("Confirm", true, " ")
	pl.MustFail(err)

	if confirmed {
		pl.Println("You confirmed")
	} else {
		pl.Println("You declined")
	}

	input, err := pl.Read("Input: ", 64)
	pl.MustFail(err)

	pl.Printf("You entered '%s'\n", input)

	key, err := pl.ReadOne("Key: ", true)
	pl.MustFail(err)

	pl.Printf("You entered '%s'\n", string(key))

	hidden, err := pl.ReadHidden("Hidden: ")
	pl.MustFail(err)

	pl.Printf("You entered '%s'\n", hidden)

	masked, err := pl.ReadMask("Masked: ", plain.MaskStar)
	pl.MustFail(err)

	pl.Printf("You entered '%s'\n", masked)

	options := []plain.SelectOption{
		colorOption{"Red", "A warm, energetic color often associated with passion and urgency."},
		colorOption{"Green", "A natural, balanced color that suggests growth and tranquility."},
		colorOption{"Blue", "A cool, calming color commonly associated with trust and stability."},
		colorOption{"Yellow", "A bright, optimistic color that evokes sunshine and creativity."},
	}

	index, err := pl.SelectWithDescription("Select: ", options)
	pl.MustFail(err)

	pl.Printf("You selected '%s'\n", options[index].Label())
}
