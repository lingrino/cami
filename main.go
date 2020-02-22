package main

import (
	"fmt"

	"github.com/lingrino/cami/cmd"
)

// version is populated at build time by goreleaser
var version = "dev"

// main is the primary entrypoint to the application
func main() {
	err := cmd.Execute(version)
	if err != nil {
		fmt.Println(err)
	}
}
