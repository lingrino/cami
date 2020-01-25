package cami_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lingrino/cami/cami"
)

func main() {
	var err error

	aws, err := cami.NewAWS(&cami.Config{DryRun: false})
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	err = aws.Auth()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	deleted, err := aws.DeleteUnusedAMIs()
	if len(deleted) == 0 && err == nil {
		fmt.Println("nothing to delete")
	}

	var eda *cami.ErrDeleteAMIs
	if err != nil {
		if errors.As(err, &eda) {
			fmt.Printf("Failed to delete:\n  %s\n", strings.Join(eda.IDs, "\n  "))
		} else {
			fmt.Printf("UNKNOWN ERROR: %v\n", err)
		}
	}
	if len(deleted) > 0 {
		fmt.Printf("Successfully deleted:\n  %s\n", strings.Join(deleted, "\n  "))
	}
}
