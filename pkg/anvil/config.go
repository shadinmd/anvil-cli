package anvil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ProjectsFilePath = "~/.config/anvil/projects.json"

func LoadProjects() ([]Project, error) {
	path := getProjectsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Project{}, nil
		}
		return nil, fmt.Errorf("failed to read projects file: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to unmarshal projects: %w", err)
	}

	for i := range projects {
		projects[i].Active = false
		projects[i].Dir = ExpandPath(projects[i].Dir)
	}

	return projects, nil
}

func SaveProjects(projects []Project) error {
	path := getProjectsPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal projects: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write projects file: %w", err)
	}

	return nil
}

func getProjectsPath() string {
	return ExpandPath(ProjectsFilePath)
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // fallback
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
