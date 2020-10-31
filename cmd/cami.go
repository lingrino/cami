package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/lingrino/cami/cami"
	"github.com/spf13/cobra"
)

const (
	flagDryRunDesc = "Set dryrun to true to run through the deletion without deleting any AMIs."
)

// camiCmd returns our root cami command.
func camiCmd() *cobra.Command {
	// dryrun determines if cami should test deletion but not actually delete the AMIs
	var dryrun bool

	cmd := &cobra.Command{
		Use:   "cami",
		Short: "cami is an API and CLI for removing unused AMIs from your AWS account.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			aws, err := cami.NewAWS(&cami.Config{DryRun: dryrun})
			if err != nil {
				log.Fatalf("ERROR: %v\n", err)
			}

			err = aws.Auth()
			if err != nil {
				log.Fatalf("ERROR: %v\n", err)
			}

			deleted, err := aws.DeleteUnusedAMIs()
			if len(deleted) == 0 && err == nil {
				fmt.Println("nothing to delete")
			}

			var eda *cami.ErrDeleteAMIs
			if err != nil {
				if errors.As(err, &eda) {
					log.Fatalf("Failed to delete:\n  %s\n", strings.Join(eda.IDs, "\n  "))
				} else {
					log.Fatalf("UNKNOWN ERROR: %v\n", err)
				}
			}
			if len(deleted) > 0 {
				fmt.Printf("Successfully deleted:\n  %s\n", strings.Join(deleted, "\n  "))
			}
		},
	}

	cmd.Flags().BoolVarP(&dryrun, "dryrun", "d", false, flagDryRunDesc)

	return cmd
}

// Execute calls the command returned by camiCmd and sets the version flag passed from main.go.
func Execute(v string) error {
	cami := camiCmd()
	cami.AddCommand(versionCmd(v))

	err := cami.Execute()
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return nil
}
