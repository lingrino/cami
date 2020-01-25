package main

import (
	"fmt"

	"github.com/lingrino/cami/cmd"
)

// version is populated at build time by goreleaser
var version = "dev"

// executeF is the function that should be used to call `cami`
var executeF = cmd.Execute

// main is the primary entrypoint to the application
func main() {
	var err error

	err = executeF(version)
	if err != nil {
		fmt.Println(err)
	}
}
