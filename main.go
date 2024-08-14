package main

import (
	"github.com/alexflint/go-arg"
	"github.com/tforceaio/tf-unifiler-go/cmd"
)

var invokeArgs cmd.Args

func main() {
	invokeArgs = cmd.Args{}
	arg.MustParse(&invokeArgs)
}
