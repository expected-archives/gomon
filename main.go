package main

import (
	"github.com/spf13/cobra"

	"github.com/expectedsh/gomon/commands/run"
)

var rootCmd = &cobra.Command{
	Use:     "gomon",
	Short:   "Gomon is a utility tool to run multiple application at once that support hot reloading.",
	Example: "gomon run",
}

func main() {
	_ = rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(run.Command)
}
