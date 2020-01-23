package main

import (
	"fmt"
	"log"

	"github.com/lingrino/cami/cami"
)

func main() {
	aws, err := cami.NewAWS(&cami.Config{DryRun: true})
	if err != nil {
		log.Fatal(err)
	}

	err = aws.Auth()
	if err != nil {
		log.Fatal(err)
	}

	deleted, err := aws.DeleteUnusedAMIs()
	fmt.Println(err)
	fmt.Println(deleted)
}
