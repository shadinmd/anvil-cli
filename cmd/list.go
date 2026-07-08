package cmd

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	var (
		activeStyle   = lipgloss.NewStyle().Foreground(ui.Color.Success)
		inactiveStyle = lipgloss.NewStyle().Foreground(ui.Color.Secondary)
		nameStyle     = lipgloss.NewStyle().Bold(true).Foreground(ui.Color.Primary)
		dirStyle      = lipgloss.NewStyle().Foreground(ui.Color.Secondary)
		titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(ui.Color.Primary).Underline(true)

		showArchived bool
		showAll      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list projects",
		Run: func(cmd *cobra.Command, args []string) {

			logger.Log("Fetching all projects")
			allProjects := anvil.GetAllProjects()

			var projects []anvil.Project
			for _, p := range allProjects {
				if showAll || (showArchived && p.Archived) || (!showArchived && !p.Archived) {
					projects = append(projects, p)
				}
			}

			logger.Logf("Found %d projects", len(projects))

			if len(projects) == 0 {
				fmt.Println("No projects found.")
				return
			}

			fmt.Println(titleStyle.Render("Projects:"))
			for _, p := range projects {
				status := inactiveStyle.Render("○")
				if p.Active {
					status = activeStyle.Render("●")
				}
				name := p.Name
				if p.Archived {
					name += " (archived)"
				}
				fmt.Printf("%s %-20s %s\n", status, nameStyle.Render(name), dirStyle.Render(p.Dir))
			}
		},
	}

	cmd.Flags().BoolVarP(&showArchived, "archived", "a", false, "list only archived projects")
	cmd.Flags().BoolVarP(&showAll, "all", "A", false, "list all projects, including archived ones")

	return cmd
}
