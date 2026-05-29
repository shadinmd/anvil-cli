package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ConfirmModel struct {
	title   string
	message string
	Done    bool
	Confirm bool
	Cancel  bool
	width   int
	height  int
}

func NewConfirmModel(title, message string) *ConfirmModel {
	return &ConfirmModel{
		title:   title,
		message: message,
	}
}

func (m *ConfirmModel) Init() tea.Cmd {
	return nil
}

func (m *ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			m.Confirm = true
			m.Done = true
		case "n", "N", "esc":
			m.Cancel = true
			m.Done = true
		}
	}
	return m, nil
}

func (m *ConfirmModel) View() tea.View {
	var s strings.Builder

	// Create the modal box
	modalWidth := 60
	if m.width < modalWidth+4 {
		modalWidth = m.width - 4
	}

	title := titleStyle.Width(modalWidth).Render(m.title)
	
	messageStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Align(lipgloss.Center).
		Padding(1, 0)
	
	message := messageStyle.Render(m.message)

	buttonStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 3).
		Margin(0, 2)

	yesButton := buttonStyle.Copy().
		Foreground(Color.Success).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Color.Success).
		Render("YES (y)")

	noButton := buttonStyle.Copy().
		Foreground(Color.Error).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Color.Error).
		Render("NO (n)")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesButton, noButton)
	buttons = lipgloss.NewStyle().Width(modalWidth).Align(lipgloss.Center).Render(buttons)

	content := lipgloss.JoinVertical(lipgloss.Center, title, message, buttons)
	
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Color.Primary).
		Padding(1, 2).
		Render(content)

	// Center the modal on screen
	vCenter := (m.height - lipgloss.Height(modal)) / 2
	if vCenter < 0 {
		vCenter = 0
	}
	
	hCenter := (m.width - lipgloss.Width(modal)) / 2
	if hCenter < 0 {
		hCenter = 0
	}

	s.WriteString(strings.Repeat("\n", vCenter))
	lines := strings.Split(modal, "\n")
	for _, line := range lines {
		s.WriteString(strings.Repeat(" ", hCenter) + line + "\n")
	}

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}
