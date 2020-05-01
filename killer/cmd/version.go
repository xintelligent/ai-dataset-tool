package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use: "version",
	Run: showVersion,
}

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Println("killer: v1.0")
}
