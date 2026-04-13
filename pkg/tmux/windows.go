package tmux

import (
	"os/exec"

	"github.com/shadinmd/anvil-cli/pkg/logger"
)

func CreateWindow(sessionName, windowName, projectDir string) {
	logger.Logf("Creating window %s in session %s", windowName, sessionName)
	exec.Command("tmux", "new-window", "-t", sessionName, "-n", windowName, "-c", projectDir).Run()
}

func RenameWindow(target, newName string) {
	logger.Logf("Renaming window %s to %s", target, newName)
	exec.Command("tmux", "rename-window", "-t", target, newName).Run()
}

func SendKeys(target, command string) {
	logger.Logf("Sending keys to %s: %s", target, command)
	exec.Command("tmux", "send-keys", "-t", target, command, "C-m").Run()
}

func SelectWindow(target string) {
	logger.Logf("Selecting window: %s", target)
	exec.Command("tmux", "select-window", "-t", target).Run()
}

func SelectLayout(target, layout string) {
	logger.Logf("Selecting layout %s for %s", layout, target)
	exec.Command("tmux", "select-layout", "-t", target, layout).Run()
}

