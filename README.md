# go.pipeline
go.pipeline is a utility library that imitates unix pipeline. It simplifies chaining unix commands (and other stuff) in Go.

## Installation
```
go get -u github.com/songgao/go.pipeline
```

## Documentation
[http://godoc.org/github.com/songgao/go.pipeline](http://godoc.org/github.com/songgao/go.pipeline)

## Examples

### Execute a single unix command
```go
package main

import "github.com/songgao/go.pipeline"

func main() {
	pipeline.StartPipelineWithCommand("uname", "-a").PrintAll()
}
```

### Chain multiple unix commands
Following code is equal to `ls -h -l /bin | egrep --color=always -e "\\ *[0-9.]*M\\ *"`, plus a counter pre-pended to each line.
```go
package main

import (
	"fmt"
	"github.com/songgao/go.pipeline"
)

func main() {
	counter := 0

	pipeline.NewPipeline().
		ChainCommand("ls", "-h", "-l", "/bin").ChainCommand("egrep", "--color=always", "-e", `\ *[0-9.]*M\ *`).
		ChainLineProcessor(func(in string) string {
		counter++
		return fmt.Sprintf("%d:\t%s", counter, in)
	}, nil).PrintAll()
}
```
An equivalent shorter version is:
```go
package main

import (
	"fmt"
	"github.com/songgao/go.pipeline"
)

func main() {
	counter := 0

	pipeline.NewPipeline().
		C("ls", "-h", "-l", "/bin").C("egrep", "--color=always", "-e", `\ *[0-9.]*M\ *`).
		L(func(in string) string {
		counter++
		return fmt.Sprintf("%d:\t%s", counter, in)
	}, nil).P()
}
```

### Colorize `go build` output
Check out `colorgo` project: [https://github.com/songgao/colorgo](https://github.com/songgao/colorgo)

## TODO
* 2>&1, tee, etc.
* package doc
* more organized README.md
