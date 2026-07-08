package cmd

import (
	"fmt"
	"strings"

	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [project]",
		Short: "Start a project by its name or ID",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projects := anvil.GetAllProjects()
			var project *anvil.Project

			if len(args) == 0 {
				var activeProjects []anvil.Project
				for _, p := range projects {
					if !p.Archived {
						activeProjects = append(activeProjects, p)
					}
				}

				if len(activeProjects) == 0 {
					fmt.Println("No active projects found to start.")
					return
				}

				var err error
				project, err = ui.RunListUI(activeProjects)
				if err != nil {
					fmt.Printf("Error showing selection UI: %v\n", err)
					return
				}
				if project == nil {
					return
				}
			} else {
				target := strings.ToLower(args[0])
				logger.Logf("Attempting to start project: %s", target)

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

				if project.Archived {
					fmt.Printf("Error: Project '%s' is archived.\nUse 'anvil unarchive %s' to restore it first.\n", project.Name, project.Id)
					return
				}
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
