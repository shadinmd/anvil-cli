package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
	"github.com/shadinmd/anvil-cli/pkg/git"
	"github.com/shadinmd/anvil-cli/pkg/tmux"
)

type WorktreeItem struct {
	Worktree git.Worktree
	Active   bool
	IsNew    bool
}

type WorktreePickerModel struct {
	project       *anvil.Project
	items         []WorktreeItem
	filteredItems []WorktreeItem
	cursor        int
	searchQuery   string
	Choice        *WorktreeItem
	width         int
	quitting      bool
}

func NewWorktreePickerModel(project *anvil.Project, worktrees []git.Worktree) *WorktreePickerModel {
	sessions, _ := tmux.GetAllSessions()
	activeSessions := make(map[string]bool)
	for _, s := range sessions {
		activeSessions[s] = true
	}

	var items []WorktreeItem
	for _, wt := range worktrees {
		sessionName := anvil.WorktreeSessionName(project, wt.Branch, wt.IsMain)
		items = append(items, WorktreeItem{
			Worktree: wt,
			Active:   activeSessions[sessionName],
		})
	}

	m := &WorktreePickerModel{
		project:  project,
		items:    items,
		width:    80,
	}
	m.filterItems()
	return m
}

func (m *WorktreePickerModel) Init() tea.Cmd {
	return nil
}

func (m *WorktreePickerModel) filterItems() {
	q := strings.TrimSpace(m.searchQuery)
	var filtered []WorktreeItem
	exactMatch := false

	for _, item := range m.items {
		if q == "" || fuzzyMatch(item.Worktree.Branch, q) || fuzzyMatch(item.Worktree.Path, q) {
			filtered = append(filtered, item)
		}
		if strings.EqualFold(item.Worktree.Branch, q) {
			exactMatch = true
		}
	}

	if q != "" && !exactMatch {
		filtered = append(filtered, WorktreeItem{
			Worktree: git.Worktree{
				Branch: q,
			},
			IsNew: true,
		})
	}

	m.filteredItems = filtered
	if m.cursor >= len(m.filteredItems) {
		m.cursor = 0
		if len(m.filteredItems) > 0 {
			m.cursor = len(m.filteredItems) - 1
		}
	}
}

func (m *WorktreePickerModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.KeyPressMsg:
		k := msg.String()
		switch k {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "ctrl+n":
			if m.cursor < len(m.filteredItems)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.filteredItems) > 0 {
				m.Choice = &m.filteredItems[m.cursor]
			}
			m.quitting = true
			return m, tea.Quit

		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterItems()
			}

		case "space":
			m.searchQuery += " "
			m.filterItems()

		default:
			if len(k) == 1 {
				m.searchQuery += k
				m.filterItems()
			}
		}
	}

	return m, nil
}

func (m *WorktreePickerModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var s strings.Builder

	prompt := lipgloss.NewStyle().Foreground(Color.Primary).Bold(true).Render("?")
	label := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf(" Switch worktree [%s]: ", m.project.Name))
	searchText := m.searchQuery
	cursor := lipgloss.NewStyle().Foreground(Color.Primary).Blink(true).Render("|")

	s.WriteString(prompt + label + searchText + cursor + "\n")

	maxResults := 10
	start := 0
	if m.cursor >= maxResults {
		start = m.cursor - maxResults + 1
	}

	for i := start; i < len(m.filteredItems) && i < start+maxResults; i++ {
		item := m.filteredItems[i]
		pointer := "  "
		style := lipgloss.NewStyle().Foreground(Color.Foreground)

		if i == m.cursor {
			pointer = lipgloss.NewStyle().Foreground(Color.Primary).Render("> ")
			style = lipgloss.NewStyle().Foreground(Color.Primary).Bold(true)
		}

		if item.IsNew {
			newStyle := lipgloss.NewStyle().Foreground(Color.Primary).Italic(true)
			s.WriteString(fmt.Sprintf("%s%s\n", pointer, newStyle.Render(fmt.Sprintf("+ Create worktree \"%s\"", item.Worktree.Branch))))
			continue
		}

		status := lipgloss.NewStyle().Foreground(Color.Secondary).Render("○")
		if item.Active {
			status = lipgloss.NewStyle().Foreground(Color.Success).Render("●")
		}

		currentTag := ""
		if item.Worktree.IsCurrent {
			currentTag = lipgloss.NewStyle().Foreground(Color.Secondary).Render(" (current)")
		}

		mainTag := ""
		if item.Worktree.IsMain {
			mainTag = lipgloss.NewStyle().Foreground(Color.Secondary).Render(" [main]")
		}

		s.WriteString(fmt.Sprintf("%s%s %s%s%s\n", pointer, status, style.Render(item.Worktree.Branch), mainTag, currentTag))
	}

	if len(m.filteredItems) == 0 {
		s.WriteString(helpStyle.Render("  No matches found") + "\n")
	}

	help := helpStyle.Render("  (type to filter, ↑/↓ to navigate, enter to switch, esc to cancel)")
	s.WriteString(help)

	return tea.NewView(s.String())
}

func RunWorktreePicker(project *anvil.Project, worktrees []git.Worktree) (*WorktreeItem, error) {
	m := NewWorktreePickerModel(project, worktrees)
	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		return nil, err
	}
	return m.Choice, nil
}
