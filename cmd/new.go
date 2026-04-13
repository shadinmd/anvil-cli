package cmd

import (
	"github.com/spf13/cobra"
)

func NewNewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	return cmd
}
