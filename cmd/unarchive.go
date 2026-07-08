package cmd

import (
	"fmt"
	"strings"

	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func NewUnarchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unarchive [project]",
		Aliases: []string{"restore"},
		Short:   "Restore an archived project",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projects := anvil.GetAllProjects()
			
			// Filter only archived projects
			var archivedProjects []anvil.Project
			for _, p := range projects {
				if p.Archived {
					archivedProjects = append(archivedProjects, p)
				}
			}

			var project *anvil.Project

			if len(args) == 0 {
				if len(archivedProjects) == 0 {
					fmt.Println("No archived projects to restore.")
					return
				}
				var err error
				project, err = ui.RunListUI(archivedProjects)
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

				if !project.Archived {
					fmt.Printf("Project '%s' is not archived.\n", project.Name)
					return
				}
			}

			if err := anvil.UnarchiveProject(project.Id); err != nil {
				fmt.Printf("Error restoring project: %v\n", err)
				return
			}

			fmt.Printf("Project '%s' has been successfully restored.\n", project.Name)
		},
	}

	return cmd
}
