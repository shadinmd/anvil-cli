package anvil

import (
	"fmt"
	"strings"

	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/tmux"
)

type Project struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Dir     string   `json:"dir"`
	Windows []Window `json:"windows"`
	Active  bool     `json:"-"`
}

type Window struct {
	Name   string `json:"name"`
	Layout string `json:"layout,omitempty"` // e.g., "tiled", "even-horizontal"
	Panes  []Pane `json:"panes"`
}

type Pane struct {
	Cmd   string `json:"cmd"`
	Split string `json:"split,omitempty"` // "h" or "v" (defaults to v)
}

// tmux.ProjectConfig implementation
func (p Project) GetId() string  { return p.Id }
func (p Project) GetDir() string { return p.Dir }
func (p Project) GetWindows() []tmux.WindowConfig {
	res := make([]tmux.WindowConfig, len(p.Windows))
	for i, w := range p.Windows {
		res[i] = w
	}
	return res
}

// tmux.WindowConfig implementation
func (w Window) GetName() string   { return w.Name }
func (w Window) GetLayout() string { return w.Layout }
func (w Window) GetPanes() []tmux.PaneConfig {
	res := make([]tmux.PaneConfig, len(w.Panes))
	for i, p := range w.Panes {
		res[i] = p
	}
	return res
}

// tmux.PaneConfig implementation
func (p Pane) GetCmd() string   { return p.Cmd }
func (p Pane) GetSplit() string { return p.Split }

func GetAllProjects() []Project {
	lookup := make(map[string]bool)
	sessions, err := tmux.GetAllSessions()
	if err != nil {
		fmt.Printf("Error loading sessions: %v\n", err)
		return []Project{}
	}

	projects, err := LoadProjects()
	if err != nil {
		fmt.Printf("Error loading projects: %v\n", err)
		return []Project{}
	}
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return []Project{}
	}

	for _, s := range sessions {
		lookup[s] = true
	}

	for i := range projects {
		projects[i].Active = lookup[projects[i].Id]
	}

	return projects
}

func UpdateProject(targetId string, name, dir string, windows []Window) error {
	logger.Logf("Updating project: %s (new name: %s, dir: %s)", targetId, name, dir)
	projects, err := LoadProjects()
	if err != nil {
		return err
	}

	foundIdx := -1
	for i, p := range projects {
		if p.Id == targetId {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		return fmt.Errorf("project with ID %s not found", targetId)
	}

	// Create new ID in case name changed
	newId := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	// Ensure new name doesn't collide with ANOTHER project
	if newId != targetId {
		for _, p := range projects {
			if p.Id == newId {
				return fmt.Errorf("project with id %s already exists", newId)
			}
		}
	}

	projects[foundIdx].Id = newId
	projects[foundIdx].Name = name
	projects[foundIdx].Dir = dir
	projects[foundIdx].Windows = windows

	return SaveProjects(projects)
}

func CreateProject(name, dir string, windows []Window) error {
	logger.Logf("Creating project: %s (dir: %s, windows: %d)", name, dir, len(windows))
	projects, err := LoadProjects()
	if err != nil {
		logger.Logf("Error loading projects: %v", err)
		return err
	}

	id := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	for _, p := range projects {
		if p.Id == id {
			logger.Logf("Project already exists with ID: %s", id)
			return fmt.Errorf("project with id %s already exists", id)
		}
	}

	project := Project{
		Id:      id,
		Name:    name,
		Dir:     dir,
		Windows: windows,
	}

	projects = append(projects, project)

	if err := SaveProjects(projects); err != nil {
		logger.Logf("Error saving projects: %v", err)
		return err
	}

	logger.Logf("Project created: %s", id)
	return nil
}

func StartProject(project *Project) error {
	logger.Logf("Starting project %s in dir %s", project.Id, project.Dir)
	return tmux.CreateSession(project)
}

func SwitchProject(project *Project) error {
	logger.Logf("Switching to project %s", project.Id)
	return tmux.AttachSession(project.Id)
}

func StopProject(project *Project) error {
	logger.Logf("Stopping project %s", project.Id)
	return tmux.KillSession(project.Id)
}

func DeleteProject(id string) error {
	logger.Logf("Deleting project: %s", id)
	projects, err := LoadProjects()
	if err != nil {
		return err
	}

	foundIdx := -1
	for i, p := range projects {
		if p.Id == id {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		return fmt.Errorf("project with ID %s not found", id)
	}

	projects = append(projects[:foundIdx], projects[foundIdx+1:]...)

	return SaveProjects(projects)
}
