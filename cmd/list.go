package cmd

import (
	"fmt"

	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all projects",
		Run: func(cmd *cobra.Command, args []string) {

			logger.Log("Fetching all projects")
			projects := anvil.GetAllProjects()
			logger.Logf("Found %d projects", len(projects))

			project, err := ui.RunListUI(projects)
			if err != nil {
				logger.Logf("Error running UI: %v", err)
				fmt.Printf("Error running UI: %v\n", err)
				return
			}

			if project != nil {
				logger.Logf("Project selected: %s (active: %t)", project.Name, project.Active)
				var err error
				if project.Active {
					logger.Logf("Switching to active project: %s", project.Id)
					err = anvil.SwitchProject(project)
				} else {
					logger.Logf("Starting project: %s", project.Id)
					err = anvil.StartProject(project)
				}

				if err != nil {
					logger.Logf("Error during start/switch: %v", err)
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				logger.Log("No project selected")
			}
		},
	}

	return cmd
}
