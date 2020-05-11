package main

import (
	"github.com/spf13/cobra"

	"github.com/expectedsh/gomon/commands/older_pids"
	"github.com/expectedsh/gomon/commands/run"
)

var rootCmd = &cobra.Command{
	Version: "0.1",
	Use:     "gomon",
	Short:   "Gomon is a utility tool to run multiple application at once that support hot reloading.",
	Example: "gomon run",
}

func main() {
	_ = rootCmd.Execute()
}

func init() {
	rootCmd.Flags().String(
		"config",
		"./.gomon.yaml",
		"the location of the config file that describe what to run")

	rootCmd.AddCommand(run.Command)
	rootCmd.AddCommand(older_pids.Command)
}
