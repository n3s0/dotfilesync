package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Print the version number of dotfilesync",
  Long:  `Print the version number of dotfilesync`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("dotfilesync v0.1")
  },
}
