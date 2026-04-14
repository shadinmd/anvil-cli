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
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all projects",
		Run: func(cmd *cobra.Command, args []string) {

			logger.Log("Fetching all projects")
			projects := anvil.GetAllProjects()
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
				fmt.Printf("%s %-20s %s\n", status, nameStyle.Render(p.Name), dirStyle.Render(p.Dir))
			}
		},
	}

	return cmd
}
