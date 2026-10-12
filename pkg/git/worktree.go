package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shadinmd/anvil-cli/pkg/logger"
)

type Worktree struct {
	Path      string
	Branch    string
	IsMain    bool
	IsCurrent bool
}

func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

func GetWorktrees(dir string, currentDir string) ([]Worktree, error) {
	cmd := exec.Command("git", "-C", dir, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	var worktrees []Worktree
	lines := strings.Split(string(out), "\n")
	var currentWt Worktree
	isFirst := true

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentWt.Path != "" {
				if isFirst {
					currentWt.IsMain = true
					isFirst = false
				}
				if currentDir != "" && (currentWt.Path == currentDir || strings.HasPrefix(currentDir, currentWt.Path+string(os.PathSeparator))) {
					currentWt.IsCurrent = true
				}
				worktrees = append(worktrees, currentWt)
				currentWt = Worktree{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "worktree":
			currentWt.Path = parts[1]
		case "branch":
			branchRef := parts[1]
			currentWt.Branch = strings.TrimPrefix(branchRef, "refs/heads/")
		case "bare":
			// skip bare
		}
	}

	if currentWt.Path != "" {
		if isFirst {
			currentWt.IsMain = true
		}
		if currentDir != "" && (currentWt.Path == currentDir || strings.HasPrefix(currentDir, currentWt.Path+string(os.PathSeparator))) {
			currentWt.IsCurrent = true
		}
		worktrees = append(worktrees, currentWt)
	}

	return worktrees, nil
}

func EnsureGitExclude(repoDir string) {
	out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--git-path", "info/exclude").Output()
	if err != nil {
		return
	}
	excludePath := strings.TrimSpace(string(out))
	if !filepath.IsAbs(excludePath) {
		excludePath = filepath.Join(repoDir, excludePath)
	}

	content, _ := os.ReadFile(excludePath)
	if !strings.Contains(string(content), ".worktrees") {
		f, err := os.OpenFile(excludePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			f.WriteString("\n.worktrees/\n")
		}
	}
}

func AddWorktree(repoDir, branch string) (string, error) {
	EnsureGitExclude(repoDir)

	targetDir := filepath.Join(repoDir, ".worktrees", branch)
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return "", err
	}

	logger.Logf("Adding git worktree for branch %s at %s", branch, targetDir)

	// Check if branch exists
	branchExists := exec.Command("git", "-C", repoDir, "show-ref", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil

	var cmd *exec.Cmd
	if branchExists {
		cmd = exec.Command("git", "-C", repoDir, "worktree", "add", targetDir, branch)
	} else {
		cmd = exec.Command("git", "-C", repoDir, "worktree", "add", "-b", branch, targetDir)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to add worktree: %s (%w)", strings.TrimSpace(string(output)), err)
	}

	return targetDir, nil
}

func RemoveWorktree(repoDir, targetPath string, force bool) error {
	logger.Logf("Removing git worktree at %s (force: %t)", targetPath, force)
	args := []string{"-C", repoDir, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, targetPath)
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func GetLatestCommit(dir string) string {
	cmd := exec.Command("git", "-C", dir, "log", "-1", "--oneline")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
