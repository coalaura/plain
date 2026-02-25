# plain

A tiny, super fast logger for Go with a couple of handy input helpers.

### Install
```bash
go get -u github.com/coalaura/plain
```

### Quick start

```golang
import (
	"github.com/coalaura/plain"
)

func main() {
	pl := plain.New()

	pl.Debugln("Hello from Debug")
	pl.Println("Hello from Print")
	pl.Warnln("Hello from Warn")
	pl.Errorln("Hello from Error")

	confirmed, _ := pl.Confirm("Confirm", true)

	if confirmed {
		pl.Println("You confirmed")
	} else {
		pl.Println("You declined")
	}

	input, _ := pl.Read("Input: ", 64)

	pl.Printf("You entered '%s'\n", input)

	options := []string{"Red", "Green", "Blue", "Yellow"}

	index, _ := pl.Select("Select: ", options)

	pl.Printf("You selected '%s'\n", options[index])

	// optional, prevents leftover un-reset colors
	pl.WaitForInterrupt(true)

	// cleanup code on exit...
}
```

### Performance tips

- Prefer `Println`/`Printf` for hot paths to avoid extra formatting work.
- Keep `WithDate("")` (default) when you do not need timestamps.
- Avoid embedding ANSI sequences in messages if color is disabled.
- Reuse a single `Plain` instance instead of creating per call.
- For high-throughput logging, direct output to buffered IO (e.g. `bufio.Writer`).
