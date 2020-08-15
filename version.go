package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version string
var Commit string

var full *bool

func init() {
	fs := VersionCmd.Flags()
	full = fs.Bool("full", false, "fully print")
}

var VersionCmd = cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(c *cobra.Command, args []string) {
		if *full {
			fmt.Printf("%v (%v)\n", Version, Commit)
		} else {
			fmt.Printf("%v\n", Version)
		}
	},
}
