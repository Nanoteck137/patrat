package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var AppName = "patrat"
var Version = "no-version"
var Commit = "no-commit"

var rootCmd = &cobra.Command{
	Use:     AppName,
	Version: Version,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func versionTemplate() string {
	return fmt.Sprintf(
		"%s: %s (%s)\n",
		AppName, Version, Commit)
}

func init() {
	rootCmd.SetVersionTemplate(versionTemplate())
}
