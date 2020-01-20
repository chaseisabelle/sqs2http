# flagz
simple go package for multiple flags with the same name/key

---
## example

The code...
```go
package main

import (
	"flag"
	"fmt"
	"github.com/chaseisabelle/flagz"
)

func main() {
	var arr flagz.Flagz

	flag.Var(&arr, "foo", "bla")

	flag.Parse()

	fmt.Printf("%+v\n", arr.Array())
}
```

Running it...
```
./main --foo=bar --foo=bar2
[bar bar2]
bar, bar2
```