# plain

A tiny, dependency-free logger for Go with a couple of handy input helpers.

### Install
```bash
go get -u github.com/coalaura/plain
```

### Quick start

```golang
import (
	"os"
	"github.com/coalaura/plain"
)

func main() {
	pl := plain.New(os.Stdout)

	pl.Println("Hello from Print")
	pl.Warnln("Hello from Warn")
	pl.Errorln("Hello from Error")

	input, err := pl.Read(os.Stdin, "Input: ", 64)
	pl.MustFail(err)

	pl.Printf("You entered '%s'\n", input)

	options := []string{"Red", "Green", "Blue", "Yellow"}

	idx, err := pl.Select("Select: ", options)
	pl.MustFail(err)

	pl.Printf("You selected '%s'\n", options[idx])
}
```
