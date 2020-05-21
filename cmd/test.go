package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

type shape struct {
	x0 int
	y0 int
	x1 int
	y1 int
}

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("一个测试命令")
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
