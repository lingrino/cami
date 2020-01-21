package main

import (
	"fmt"
	"log"

	"github.com/lingrino/cami/cami"
)

func main() {
	aws := cami.NewAWS()
	aws.Auth()

	amis, err := aws.AMIs()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(amis))

	ec2s, err := aws.EC2s(amis)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(ec2s))

	amis, err = aws.FilterAMIs(amis, ec2s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(amis))
}
