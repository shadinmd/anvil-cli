package cmd

import (
	"fmt"
	"strings"

	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func NewArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive [project]",
		Short: "Archive a project so it does not show in the default list",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projects := anvil.GetAllProjects()
			
			// Filter only non-archived projects
			var activeProjects []anvil.Project
			for _, p := range projects {
				if !p.Archived {
					activeProjects = append(activeProjects, p)
				}
			}

			var project *anvil.Project

			if len(args) == 0 {
				if len(activeProjects) == 0 {
					fmt.Println("No active projects to archive.")
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
					fmt.Printf("Project '%s' is already archived.\n", project.Name)
					return
				}
			}

			// Stop project if active
			if project.Active {
				fmt.Printf("Stopping active project '%s' before archiving...\n", project.Name)
				if err := anvil.StopProject(project); err != nil {
					fmt.Printf("Warning: failed to stop project: %v\n", err)
				}
			}

			if err := anvil.ArchiveProject(project.Id); err != nil {
				fmt.Printf("Error archiving project: %v\n", err)
				return
			}

			fmt.Printf("Project '%s' has been successfully archived.\n", project.Name)
		},
	}

	return cmd
}
