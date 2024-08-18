package cmd

import (
	"log"

	"github.com/kr/pretty"
	"github.com/nanoteck137/sewaddle-core/library"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use: "import",
	Run: func(cmd *cobra.Command, args []string) {
		lib, err := library.ReadFromDir("/Volumes/media/manga")
		if err != nil {
			log.Fatal(err)
		}

		pretty.Println(lib)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
