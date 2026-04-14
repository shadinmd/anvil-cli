package cmd

import (
	"fmt"

	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "anvil",
		Short: "A CLI tool for easy project managemnt",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if err := logger.Init(); err != nil {
				fmt.Printf("Warning: failed to initialize logger: %v\n", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.RunDashboard(); err != nil {
				fmt.Printf("Error running dashboard: %v\n", err)
			}
		},
	}

	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewStartCmd())
	cmd.AddCommand(NewNewCommand())

	return cmd
}
