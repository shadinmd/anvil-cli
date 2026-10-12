package cmd

import (
	"fmt"
	"os"

	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/git"
	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/ui"
	"github.com/spf13/cobra"
)

func NewWorktreeCmd() *cobra.Command {
	var inlineMode bool

	cmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"wt"},
		Short:   "Manage and switch worktrees of the current active project",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Log("Running worktree command")

			project, err := anvil.DetectCurrentProject()
			if err != nil {
				// Fall back to prompting for project
				projects := anvil.GetAllProjects()
				var activeProjects []anvil.Project
				for _, p := range projects {
					if !p.Archived {
						activeProjects = append(activeProjects, p)
					}
				}

				if len(activeProjects) == 0 {
					fmt.Println("No active projects found.")
					return
				}

				project, err = ui.RunListUI(activeProjects)
				if err != nil || project == nil {
					return
				}
			}

			projectDir := anvil.ExpandPath(project.Dir)
			if !git.IsGitRepo(projectDir) {
				fmt.Printf("Error: project '%s' directory (%s) is not a git repository.\n", project.Name, projectDir)
				return
			}

			if inlineMode {
				cwd, _ := os.Getwd()
				cwd = anvil.ExpandPath(cwd)

				worktrees, err := git.GetWorktrees(projectDir, cwd)
				if err != nil {
					fmt.Printf("Error getting worktrees: %v\n", err)
					return
				}

				selected, err := ui.RunWorktreePicker(project, worktrees)
				if err != nil {
					fmt.Printf("Error selecting worktree: %v\n", err)
					return
				}
				if selected == nil {
					return
				}

				var targetDir string
				isMain := selected.Worktree.IsMain

				if selected.IsNew {
					fmt.Printf("Creating worktree for branch '%s'...\n", selected.Worktree.Branch)
					targetDir, err = git.AddWorktree(projectDir, selected.Worktree.Branch)
					if err != nil {
						fmt.Printf("Error creating worktree: %v\n", err)
						return
					}
					isMain = false
				} else {
					targetDir = selected.Worktree.Path
				}

				if err := anvil.SwitchToWorktree(project, selected.Worktree.Branch, targetDir, isMain); err != nil {
					fmt.Printf("Error switching to worktree session: %v\n", err)
					return
				}
				return
			}

			if err := ui.RunWorktreeDashboard(project); err != nil {
				fmt.Printf("Error running worktree dashboard: %v\n", err)
			}
		},
	}

	cmd.Flags().BoolVarP(&inlineMode, "inline", "i", false, "Use quick inline list picker instead of full dashboard")

	return cmd
}
