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
	var stringz flagz.Flagz
	var boolz flagz.Flagz
	var intz flagz.Flagz
	var floatz flagz.Flagz

	flag.Var(&stringz, "string", "strings")
	flag.Var(&boolz, "bool", "bools")
	flag.Var(&intz, "int", "ints")
	flag.Var(&floatz, "float", "floats")

	flag.Parse()

	strings := stringz.Stringz()
	bools, boolErr := boolz.Boolz()
	ints, intErr := intz.Intz()
	floats, floatErr := floatz.Floatz()


	fmt.Printf("%+v\n", strings)
	fmt.Printf("%+v\n%+v\n", bools, boolErr)
	fmt.Printf("%+v\n%+v\n", ints, intErr)
	fmt.Printf("%+v\n%+v\n", floats, floatErr)
}

```

Running it...
```
./main --string=foo --string=bar --bool=1 --bool=0 --int=420 --int=69 --float=123.456 --float=654.321 
[foo bar]
[true false]
<nil>
[420 69]
<nil>
[123.456 654.321]
<nil>
```