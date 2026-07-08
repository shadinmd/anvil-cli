package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func NewNewCommand() *cobra.Command {
	var templateName string

	labelStyle := lipgloss.NewStyle().Foreground(ui.Color.Primary).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(ui.Color.Foreground)

	cmd := &cobra.Command{
		Use:   "new [project-name] [template-project-name]",
		Short: "Create a new project, optionally from a template",
		Args:  cobra.MaximumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var newName string

			// If no project name is provided, show the interactive picker
			if len(args) == 0 && templateName == "" {
				projects := anvil.GetAllProjects()
				var activeProjects []anvil.Project
				for _, p := range projects {
					if !p.Archived {
						activeProjects = append(activeProjects, p)
					}
				}

				if len(activeProjects) == 0 {
					// We can still create a blank project, so we don't return early here,
					// but we won't show the template picker.
					fmt.Println(labelStyle.Render("No active projects available to use as templates. Using [Blank Project]"))
				} else {
					selected, isCancelled, err := ui.RunTemplatePicker(activeProjects)
					if err != nil {
						fmt.Printf("Error picking template: %v\n", err)
						return
					}

					if isCancelled {
						fmt.Println("Cancelled.")
						return
					}

					if selected != nil {
						templateName = selected.Id
						fmt.Printf("%s %s\n", labelStyle.Render("Using template:"), inputStyle.Bold(true).Render(selected.Name))
					} else {
						fmt.Println(labelStyle.Render("Using template: [Blank Project]"))
					}
				}

				// Now ask for the new project name with style
				fmt.Print(labelStyle.Render("Enter name for the new project: "))

				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				newName = strings.TrimSpace(input)

				if newName == "" {
					fmt.Println("Project name is required. Cancelled.")
					return
				}
				fmt.Println(inputStyle.Render("Creating project: " + newName))
			} else if len(args) > 0 {
				newName = args[0]
				if len(args) > 1 {
					templateName = args[1]
				}
			} else {
				fmt.Println("Error: Project name is required.")
				return
			}

			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				return
			}
			var windows []anvil.Window

			if templateName != "" {
				logger.Logf("Creating project '%s' from template '%s'", newName, templateName)
				projects := anvil.GetAllProjects()
				var template *anvil.Project

				for i := range projects {
					if strings.ToLower(projects[i].Id) == strings.ToLower(templateName) ||
						strings.ToLower(projects[i].Name) == strings.ToLower(templateName) {
						template = &projects[i]
						break
					}
				}

				if template == nil {
					fmt.Printf("Error: Template project '%s' not found.\n", templateName)
					return
				}

				if template.Archived {
					fmt.Printf("Error: Template project '%s' is archived. Restore it first to use it as a template.\n", template.Name)
					return
				}

				windows = template.Windows
				fmt.Printf("Creating project '%s' using template '%s'...\n", newName, template.Name)
			} else {
				logger.Logf("Creating blank project '%s'", newName)
				windows = []anvil.Window{
					{
						Name: "",
						Panes: []anvil.Pane{
							{Cmd: ""},
						},
					},
				}
				fmt.Printf("Creating new blank project '%s'...\n", newName)
			}

			if err := anvil.CreateProject(newName, cwd, windows); err != nil {
				fmt.Printf("Error creating project: %v\n", err)
				return
			}

			fmt.Printf("Project '%s' created successfully in %s\n", newName, cwd)
		},
	}

	cmd.Flags().StringVarP(&templateName, "template", "t", "", "template project name or ID")

	return cmd
}
