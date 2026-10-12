package ui

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/git"
	"github.com/shadinmd/anvil-cli/pkg/logger"
	"github.com/shadinmd/anvil-cli/pkg/tmux"
)

type WorktreeDashboardModel struct {
	project     *anvil.Project
	worktrees   []git.Worktree
	filtered    []git.Worktree
	cursor      int
	offset      int
	width       int
	height      int
	searchQuery string
	err         error
	confirm     *ConfirmModel
	input       *InputModel
	quitting    bool
}

func NewWorktreeDashboardModel(project *anvil.Project) *WorktreeDashboardModel {
	m := &WorktreeDashboardModel{
		project: project,
		width:   80,
		height:  24,
	}
	m.reload()
	return m
}

func (m *WorktreeDashboardModel) reload() {
	cwd, _ := os.Getwd()
	cwd = anvil.ExpandPath(cwd)
	projectDir := anvil.ExpandPath(m.project.Dir)

	wts, err := git.GetWorktrees(projectDir, cwd)
	if err != nil {
		m.err = err
		return
	}

	m.worktrees = wts
	m.filterWorktrees()
}

func (m *WorktreeDashboardModel) filterWorktrees() {
	q := strings.ToLower(strings.TrimSpace(m.searchQuery))
	if q == "" {
		m.filtered = m.worktrees
	} else {
		var res []git.Worktree
		for _, wt := range m.worktrees {
			if fuzzyMatch(wt.Branch, q) || fuzzyMatch(wt.Path, q) {
				res = append(res, wt)
			}
		}
		m.filtered = res
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = 0
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
		}
	}
	m.scrollToCursor()
}

func (m *WorktreeDashboardModel) scrollToCursor() {
	headerHeight := 2
	footerHeight := 3
	if m.err != nil {
		footerHeight = 4
	}
	mainHeight := m.height - headerHeight - footerHeight
	if mainHeight < 5 {
		mainHeight = 5
	}
	pageSize := mainHeight - 5
	if pageSize < 1 {
		pageSize = 1
	}

	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+pageSize {
		m.offset = m.cursor - pageSize + 1
	}
}

func (m *WorktreeDashboardModel) isSessionActive(wt git.Worktree) bool {
	sessionName := anvil.WorktreeSessionName(m.project, wt.Branch, wt.IsMain)
	sessions, _ := tmux.GetAllSessions()
	for _, s := range sessions {
		if s == sessionName {
			return true
		}
	}
	return false
}

func (m *WorktreeDashboardModel) Init() tea.Cmd {
	return nil
}

func (m *WorktreeDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.confirm != nil {
		newConfirm, cmd := m.confirm.Update(msg)
		m.confirm = newConfirm.(*ConfirmModel)
		if m.confirm.Done {
			if m.confirm.Confirm && len(m.filtered) > 0 {
				wt := m.filtered[m.cursor]
				sessionName := anvil.WorktreeSessionName(m.project, wt.Branch, wt.IsMain)
				_ = tmux.KillSession(sessionName)

				projectDir := anvil.ExpandPath(m.project.Dir)
				err := git.RemoveWorktree(projectDir, wt.Path, false)
				if err != nil {
					// Retry with force
					err = git.RemoveWorktree(projectDir, wt.Path, true)
				}

				if err != nil {
					m.err = err
					logger.Logf("Error deleting worktree: %v", err)
				} else {
					m.err = nil
					m.reload()
				}
			}
			m.confirm = nil
		}
		return m, cmd
	}

	if m.input != nil {
		newInput, cmd := m.input.Update(msg)
		m.input = newInput.(*InputModel)
		if m.input.Done {
			if !m.input.Cancel && m.input.Value() != "" {
				branch := m.input.Value()
				projectDir := anvil.ExpandPath(m.project.Dir)
				_, err := git.AddWorktree(projectDir, branch)
				if err != nil {
					m.err = err
					logger.Logf("Error adding worktree: %v", err)
				} else {
					m.err = nil
					m.searchQuery = ""
					m.reload()
					// Select the new worktree
					for i, wt := range m.filtered {
						if wt.Branch == branch {
							m.cursor = i
							break
						}
					}
					m.scrollToCursor()
				}
			}
			m.input = nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.scrollToCursor()

	case tea.KeyPressMsg:
		k := msg.String()

		switch k {
		case "ctrl+c", "ctrl+q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
				m.scrollToCursor()
			}

		case "down", "ctrl+n":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.scrollToCursor()
			}

		case "ctrl+a":
			m.input = NewInputModel("New Worktree", fmt.Sprintf("Enter branch name for %s:", m.project.Name), "branch-name")

		case "ctrl+d":
			if len(m.filtered) > 0 {
				wt := m.filtered[m.cursor]
				if wt.IsMain {
					m.err = fmt.Errorf("cannot delete the main worktree")
				} else {
					m.err = nil
					m.confirm = NewConfirmModel("Delete Worktree", fmt.Sprintf("Delete worktree '%s' and remove directory?", wt.Branch))
				}
			}

		case "enter":
			if len(m.filtered) > 0 {
				wt := m.filtered[m.cursor]
				err := anvil.SwitchToWorktree(m.project, wt.Branch, wt.Path, wt.IsMain)
				if err != nil {
					m.err = err
					logger.Logf("Worktree switch error: %v", err)
				} else {
					m.quitting = true
					return m, tea.Quit
				}
			}

		case "ctrl+s":
			if len(m.filtered) > 0 {
				wt := m.filtered[m.cursor]
				sessionName := anvil.WorktreeSessionName(m.project, wt.Branch, wt.IsMain)
				_ = tmux.KillSession(sessionName)
				m.reload()
			}

		case "ctrl+r":
			m.err = nil
			m.reload()

		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterWorktrees()
			}

		case "space":
			m.searchQuery += " "
			m.filterWorktrees()

		default:
			if len(k) == 1 {
				m.searchQuery += k
				m.filterWorktrees()
			}
		}
	}

	return m, nil
}

func (m *WorktreeDashboardModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.width == 0 {
		return tea.NewView("Loading...")
	}

	if m.confirm != nil {
		m.confirm.width = m.width
		m.confirm.height = m.height
		return m.confirm.View()
	}

	if m.input != nil {
		m.input.width = m.width
		m.input.height = m.height
		return m.input.View()
	}

	var s strings.Builder

	headerHeight := 2
	footerHeight := 3
	if m.err != nil {
		footerHeight = 4
	}
	sidebarWidth := 28
	contentWidth := m.width - sidebarWidth - 1
	mainHeight := m.height - headerHeight - footerHeight
	if mainHeight < 5 {
		mainHeight = 5
	}

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Color.Primary).
		Align(lipgloss.Center).
		Width(m.width)

	s.WriteString(headerStyle.Render(fmt.Sprintf("ANVIL - WORKTREES (%s)", strings.ToUpper(m.project.Name))) + "\n\n")

	// Search box
	searchBoxStyle := focusedBorderStyle
	searchText := m.searchQuery
	if searchText == "" {
		searchText = helpStyle.Render("Filter worktrees...")
	} else {
		searchText += lipgloss.NewStyle().Foreground(Color.Primary).Blink(true).Render("|")
	}
	searchBox := searchBoxStyle.Width(sidebarWidth).Render(searchText)

	// Sidebar List
	pageSize := mainHeight - 5
	if pageSize < 1 {
		pageSize = 1
	}

	var sidebarLines []string
	for i := m.offset; i < len(m.filtered) && i < m.offset+pageSize; i++ {
		wt := m.filtered[i]
		active := m.isSessionActive(wt)

		status := lipgloss.NewStyle().Foreground(Color.Secondary).Render("○")
		if active {
			status = lipgloss.NewStyle().Foreground(Color.Success).Render("●")
		}

		branchLabel := wt.Branch
		if wt.IsMain {
			branchLabel += " [main]"
		}

		line := fmt.Sprintf("%s %s", status, branchLabel)
		if i == m.cursor {
			sidebarLines = append(sidebarLines, selectedSidebarItemStyle.Render("> "+line))
		} else {
			sidebarLines = append(sidebarLines, sidebarItemStyle.Render("  "+line))
		}
	}

	listView := borderStyle.
		Width(sidebarWidth).
		Height(mainHeight - 3).
		Render(strings.Join(sidebarLines, "\n"))

	sidebarView := lipgloss.JoinVertical(lipgloss.Left, searchBox, listView)

	// Content Area
	var content string
	if len(m.filtered) > 0 {
		wt := m.filtered[m.cursor]
		active := m.isSessionActive(wt)

		statusText := lipgloss.NewStyle().Foreground(Color.Secondary).Render("Inactive")
		if active {
			statusText = lipgloss.NewStyle().Foreground(Color.Success).Bold(true).Render("Active")
		}

		sessionName := anvil.WorktreeSessionName(m.project, wt.Branch, wt.IsMain)

		var cb strings.Builder
		cb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Branch:"), wt.Branch))
		cb.WriteString(fmt.Sprintf("%s%s\n", labelStyle.Render("Session:"), sessionName))
		cb.WriteString(fmt.Sprintf("%s   %s\n", labelStyle.Render("Path:"), wt.Path))
		cb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status:"), statusText))

		commit := git.GetLatestCommit(wt.Path)
		if commit != "" {
			cb.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Commit:"), helpStyle.Render(commit)))
		}
		cb.WriteString("\n")

		if len(m.project.Windows) > 0 {
			cb.WriteString(labelStyle.Render("Inherited Windows & Panes:") + "\n")
			for i, w := range m.project.Windows {
				isLastWindow := i == len(m.project.Windows)-1
				windowPrefix := "├──"
				if isLastWindow {
					windowPrefix = "└──"
				}

				layout := w.Layout
				if layout == "" {
					layout = "default"
				}

				windowLine := fmt.Sprintf(" %s %s [%s]", windowPrefix, w.Name, helpStyle.Render(layout))
				cb.WriteString(windowLine + "\n")

				for j, pane := range w.Panes {
					isLastPane := j == len(w.Panes)-1
					var panePrefix string
					if isLastWindow {
						panePrefix = "    "
					} else {
						panePrefix = "│   "
					}

					if isLastPane {
						panePrefix += "└──"
					} else {
						panePrefix += "├──"
					}

					cmd := pane.Cmd
					if cmd == "" {
						cmd = "default shell"
					}

					paneLine := fmt.Sprintf(" %s %s", panePrefix, cmd)
					cb.WriteString(paneLine + "\n")
				}
			}
			cb.WriteString("\n")
		}

		cb.WriteString(helpStyle.Render("Press [enter] to switch to this worktree."))
		content = cb.String()
	} else {
		content = "No worktrees found.\nPress [ctrl+a] to create a new worktree."
	}

	contentView := borderStyle.
		Width(contentWidth).
		Height(mainHeight).
		Render(content)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, contentView)
	s.WriteString(mainView + "\n")

	// Footer / Help
	helpLine1 := "[ctrl+n/p] Nav • [enter] Switch • [ctrl+s] Stop • [ctrl+a] New WT • [ctrl+d] Delete WT"
	helpLine2 := "[ctrl+r] Refresh • [ctrl+q/esc] Quit"

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("!! Error: %v", m.err)) + "\n")
	}
	s.WriteString(statusStyle.Render(helpLine1) + "\n")
	s.WriteString(statusStyle.Render(helpLine2))

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}

func RunWorktreeDashboard(project *anvil.Project) error {
	m := NewWorktreeDashboardModel(project)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
