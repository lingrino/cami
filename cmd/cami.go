package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lingrino/cami/cami"
	"github.com/spf13/cobra"
)

// version is populated at build time by goreleaser
var version = "dev"

// CamiCmd is the cobra representation of the cami command and its metadata
var CamiCmd = &cobra.Command{
	Use:   "cami",
	Short: "cami is an API and CLI for removing unused AMIs from your AWS account.",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		aws, err := cami.NewAWS(&cami.Config{DryRun: dryrun})
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
	},
}

// init is where we set all flags
func init() {
	CamiCmd.Flags().BoolVarP(&dryrun, "dryrun", "d", false, "Set dryrun to true to run through the deletion without deleting any AMIs.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the CamiCmd.
func Execute(v string) error {
	var err error

	version = v

	if CamiCmd != nil {
		err = CamiCmd.Execute()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("CamiCmd is nil")
	}

	return err
}
