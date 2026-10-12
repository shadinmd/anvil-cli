package anvil

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/tmux"
)

func DetectCurrentProject() (*Project, error) {
	projects := GetAllProjects()
	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found")
	}

	// 1. Try detecting from tmux session
	if os.Getenv("TMUX") != "" {
		out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
		if err == nil {
			currentSession := strings.TrimSpace(string(out))
			for i := range projects {
				if projects[i].Id == currentSession || strings.HasPrefix(currentSession, projects[i].Id+"--") {
					logger.Logf("Detected project from tmux session %s: %s", currentSession, projects[i].Name)
					return &projects[i], nil
				}
			}
		}
	}

	// 2. Try detecting from current working directory
	cwd, err := os.Getwd()
	if err == nil {
		cwd = ExpandPath(cwd)
		var bestMatch *Project
		bestMatchLen := 0

		for i := range projects {
			pDir := ExpandPath(projects[i].Dir)
			if cwd == pDir || strings.HasPrefix(cwd, pDir+string(os.PathSeparator)) {
				if len(pDir) > bestMatchLen {
					bestMatch = &projects[i]
					bestMatchLen = len(pDir)
				}
			}
		}

		if bestMatch != nil {
			logger.Logf("Detected project from working directory %s: %s", cwd, bestMatch.Name)
			return bestMatch, nil
		}
	}

	return nil, fmt.Errorf("could not detect active project from current directory or tmux session")
}

func SanitizeSessionName(name string) string {
	r := strings.NewReplacer("/", "-", ".", "-", ":", "-", " ", "-")
	return r.Replace(name)
}

func WorktreeSessionName(parent *Project, branch string, isMain bool) string {
	if isMain {
		return parent.Id
	}
	return fmt.Sprintf("%s--%s", parent.Id, SanitizeSessionName(branch))
}

func SwitchToWorktree(parent *Project, branch, dir string, isMain bool) error {
	sessionName := WorktreeSessionName(parent, branch, isMain)
	logger.Logf("Switching to worktree session: %s (dir: %s)", sessionName, dir)

	wtProject := Project{
		Id:      sessionName,
		Name:    fmt.Sprintf("%s (%s)", parent.Name, branch),
		Dir:     dir,
		Windows: parent.Windows,
	}

	return tmux.CreateSession(wtProject)
}
