package tmux

import (
	"os/exec"

	"github.com/shadinmd/anvil-cli/pkg/logger"
)

func SplitWindow(target, direction, projectDir string) {
	logger.Logf("Splitting window %s in direction %s", target, direction)
	args := []string{"split-window", "-t", target, "-c", projectDir}
	switch direction {
	case "h":
		args = append(args, "-h")
	case "v":
		args = append(args, "-v")
	}
	exec.Command("tmux", args...).Run()
}
