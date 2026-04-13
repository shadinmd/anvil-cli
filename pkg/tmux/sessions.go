package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shadinmd/anvil-cli/pkg/logger"
)

func GetAllSessions() ([]string, error) {
	logger.Log("Listing all tmux sessions")
	output, err := exec.Command("tmux", "list-sessions", "-F", "#S").Output()
	if err != nil {
		logger.Log("No active tmux sessions found")
		return []string{}, nil
	}

	sessionsString := strings.TrimSpace(string(output))
	if sessionsString == "" {
		return []string{}, nil
	}

	sessions := strings.Split(sessionsString, "\n")
	logger.Logf("Active sessions: %v", sessions)
	return sessions, nil
}

type ProjectConfig interface {
	GetId() string
	GetDir() string
	GetWindows() []WindowConfig
}

type WindowConfig interface {
	GetName() string
	GetLayout() string
	GetPanes() []PaneConfig
}

type PaneConfig interface {
	GetCmd() string
	GetSplit() string
}

func CreateSession(project ProjectConfig) error {
	sessionName := project.GetId()
	projectDir := project.GetDir()

	logger.Logf("Request to create session: %s in %s", sessionName, projectDir)
	if sessionExists(sessionName) {
		logger.Logf("Session %s already exists, attaching...", sessionName)
		return AttachSession(sessionName)
	}

	logger.Logf("Creating new session: %s", sessionName)
	if err := NewSession(sessionName, projectDir); err != nil {
		logger.Logf("Error creating new session: %v", err)
		return err
	}

	logger.Logf("Setting up workspace for session: %s", sessionName)
	setupWorkspace(project)

	SelectWindow(sessionName + ":1")
	return AttachSession(sessionName)
}

func NewSession(sessionName, projectDir string) error {
	logger.Logf("Executing tmux new-session -d -s %s -c %s", sessionName, projectDir)
	return exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", projectDir).Run()
}

func AttachSession(sessionName string) error {
	inTmux := os.Getenv("TMUX") != ""
	logger.Logf("Attaching to session: %s (inside tmux: %t)", sessionName, inTmux)

	var cmd *exec.Cmd
	if inTmux {
		cmd = exec.Command("tmux", "switch-client", "-t", sessionName)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", sessionName)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func KillSession(sessionName string) error {
	logger.Logf("Killing session: %s", sessionName)
	return exec.Command("tmux", "kill-session", "-t", sessionName).Run()
}

func setupWorkspace(project ProjectConfig) {
	sessionName := project.GetId()
	projectDir := project.GetDir()
	windows := project.GetWindows()

	if len(windows) == 0 {
		// Default if no windows defined
		RenameWindow(sessionName+":1", "Editor")
		SendKeys(sessionName+":1", "nvim .")
		return
	}

	for i, w := range windows {
		windowId := fmt.Sprintf("%s:%d", sessionName, i+1)
		if i == 0 {
			RenameWindow(windowId, w.GetName())
		} else {
			CreateWindow(sessionName, w.GetName(), projectDir)
		}

		for j, p := range w.GetPanes() {
			if j > 0 {
				SplitWindow(windowId, p.GetSplit(), projectDir)
			}
			// When splitting, the new pane becomes the active one.
			// Sending keys to the windowId targets the active pane in that window.
			if p.GetCmd() != "" {
				SendKeys(windowId, p.GetCmd())
			}
		}

		if w.GetLayout() != "" {
			layout := w.GetLayout()
			switch layout {
			case "horizontal":
				layout = "even-horizontal"
			case "vertical":
				layout = "even-vertical"
			}
			SelectLayout(windowId, layout)
		}
	}
}

func generateSessionName() string {
	sessions, _ := GetAllSessions()
	sessionNumber := 1
	for {
		name := fmt.Sprintf("session-%d", sessionNumber)
		found := false
		for _, s := range sessions {
			if s == name {
				found = true
				break
			}
		}
		if !found {
			return name
		}
		sessionNumber++
	}
}

func sessionExists(sessionName string) bool {
	sessions, _ := GetAllSessions()
	for _, s := range sessions {
		if s == sessionName {
			return true
		}
	}
	return false
}
