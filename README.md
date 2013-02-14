# go.pipeline
go.pipeline is a utility library that imitates unix pipeline. It simplifies chaining unix commands (and other stuff) in Go.

## Installation
```
go get -u github.com/songgao/go.pipeline
```

## Documentation
[http://godoc.org/github.com/songgao/go.pipeline](http://godoc.org/github.com/songgao/go.pipeline)

## Examples
Following code is equal to `find ~/src | egrep --color=always -e "\\.go$"`, plus a counter pre-pended to each line.
```go
package main

import (
	"github.com/songgao/go.pipeline"
	"os"
    "fmt"
)

func main() {
	p := pipeline.NewPipeline()

	counter := 0
    p.
    ChainCommand("find", os.Getenv("HOME")+"/src/").
    ChainCommand("egrep", "--color=always", "-e", "\\.go$").
    ChainLineProcessor(func(in string) string {
		counter++
		return fmt.Sprintf("%d\t%s", counter, in)
	}, nil).
    PrintAll()
}
```
A equivalent shorter version is:
```go
package main

import (
	"github.com/songgao/go.pipeline"
	"os"
    "fmt"
)

func main() {
	counter := 0

	pipeline.NewPipeline().
    C("find", os.Getenv("HOME")+"/src/").
    C("egrep", "--color=always", "-e", "\\.go$").
    L(func(in string) string {
		counter++
		return fmt.Sprintf("%d\t%s", counter, in)
	}, nil).
    P()
}
```
## TODO
* 2>&1, tee, etc.
* package doc
* more organized README.md
