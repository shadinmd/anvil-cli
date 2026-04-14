package cmd

import (
	"fmt"
	"strings"

	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [project]",
		Short: "Start a project by its name or ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			target := strings.ToLower(args[0])
			logger.Logf("Attempting to start project: %s", target)

			projects := anvil.GetAllProjects()
			var project *anvil.Project

			for i := range projects {
				if strings.ToLower(projects[i].Id) == target || strings.ToLower(projects[i].Name) == target {
					project = &projects[i]
					break
				}
			}

			if project == nil {
				fmt.Printf("Error: Project '%s' not found.\n", target)
				return
			}

			if project.Active {
				fmt.Printf("Project '%s' is already active. Switching to it...\n", project.Name)
				if err := anvil.SwitchProject(project); err != nil {
					fmt.Printf("Error switching to project: %v\n", err)
				}
				return
			}

			fmt.Printf("Starting project '%s'...\n", project.Name)
			if err := anvil.StartProject(project); err != nil {
				fmt.Printf("Error starting project: %v\n", err)
			}
		},
	}

	return cmd
}
