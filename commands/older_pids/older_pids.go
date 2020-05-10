package older_pids

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/expectedsh/gomon/pkg/pids"
	"github.com/expectedsh/gomon/pkg/utils"
)

var Command = &cobra.Command{
	Use:          "older-pids",
	Short:        "Get pids running",
	Example:      "gomon run",
	RunE:         run,
	SilenceUsage: true,
}

func run(c *cobra.Command, _ []string) error {
	cfg, err := c.Root().Flags().GetString("config")
	if err != nil {
		return err
	}

	cfgHash := ""

	if err := utils.InitConfigHash(cfg, &cfgHash); err != nil {
		return err
	}

	load, err := pids.Load(cfgHash)
	if err != nil {
		return err
	}

	for app, pid := range load {
		fmt.Println(app+":", pid)
	}

	return nil
}
