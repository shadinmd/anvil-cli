package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/shadinmd/anvil-cli/pkg/anvil"
)

type FormFieldType int

const (
	FieldProjectName FormFieldType = iota
	FieldProjectDir
	FieldWindowName
	FieldWindowLayout
	FieldPaneCmd
)

type FormField struct {
	fieldType  FormFieldType
	label      string
	value      string
	suggestion string
	isChoice   bool
	choices    []string
	choiceIdx  int
	windowIdx  int
}

type FormModel struct {
	fields   []FormField
	cursor   int
	offset   int
	width    int
	height   int
	active   bool
	Done     bool
	Cancel   bool
	isEdit   bool
	targetId string
}

func NewNewProjectForm() *FormModel {
	m := &FormModel{}
	m.fields = []FormField{
		{fieldType: FieldProjectName, label: "Project Name", value: ""},
		{fieldType: FieldProjectDir, label: "Project Directory", value: ""},
	}
	m.addWindow("Main")
	return m
}

func NewEditProjectForm(p anvil.Project) *FormModel {
	m := &FormModel{
		isEdit:   true,
		targetId: p.Id,
	}
	m.fields = []FormField{
		{fieldType: FieldProjectName, label: "Project Name", value: p.Name},
		{fieldType: FieldProjectDir, label: "Project Directory", value: p.Dir},
	}

	for i, w := range p.Windows {
		winIdx := i
		m.fields = append(m.fields,
			FormField{fieldType: FieldWindowName, label: fmt.Sprintf("Window %d Name", winIdx+1), value: w.Name, windowIdx: winIdx},
			FormField{fieldType: FieldWindowLayout, label: "Layout", isChoice: true, choices: []string{"horizontal", "vertical", "tiled"}, windowIdx: winIdx},
		)

		lastField := &m.fields[len(m.fields)-1]
		for idx, c := range lastField.choices {
			if c == w.Layout {
				lastField.choiceIdx = idx
				break
			}
		}

		for j, pane := range w.Panes {
			m.fields = append(m.fields, FormField{
				fieldType: FieldPaneCmd,
				label:     fmt.Sprintf("Pane %d", j+1),
				value:     pane.Cmd,
				windowIdx: winIdx,
			})
		}
	}
	return m
}

func (m *FormModel) addWindow(name string) {
	winIdx := 0
	for _, f := range m.fields {
		if f.fieldType == FieldWindowName && f.windowIdx >= winIdx {
			winIdx = f.windowIdx + 1
		}
	}

	m.fields = append(m.fields,
		FormField{fieldType: FieldWindowName, label: fmt.Sprintf("Window %d Name", winIdx+1), value: "", windowIdx: winIdx},
		FormField{fieldType: FieldWindowLayout, label: "Layout", isChoice: true, choices: []string{"horizontal", "vertical", "tiled"}, windowIdx: winIdx},
		FormField{fieldType: FieldPaneCmd, label: "Pane 1", value: "", windowIdx: winIdx},
	)
}

func (m *FormModel) scrollToCursor() {
	if m.height <= 10 {
		return
	}

	// Each field takes roughly 3 lines (label + input + spacing)
	// Calculate how many fields can fit in the available height
	availableHeight := m.height - 12 // Account for title, help text, and margins
	pageSize := availableHeight / 3
	if pageSize < 1 {
		pageSize = 1
	}

	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+pageSize {
		m.offset = m.cursor - pageSize + 1
	}
}

func (m *FormModel) Init() tea.Cmd {
	return nil
}

func (m *FormModel) isComplete() bool {
	for _, f := range m.fields {
		if !f.isChoice && f.value == "" {
			return false
		}
	}
	return true
}

func (m *FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		field := &m.fields[m.cursor]
		key := msg.String()

		switch key {
		case "esc":
			m.Cancel = true
			return m, nil
		case "tab", "right":
			if field.suggestion != "" && field.fieldType == FieldProjectDir {
				field.value = field.suggestion
				field.suggestion = ""
				return m, nil
			}
			m.cursor = (m.cursor + 1) % len(m.fields)
			m.scrollToCursor()
			return m, nil
		case "down":
			m.cursor = (m.cursor + 1) % len(m.fields)
			m.scrollToCursor()
			return m, nil
		case "shift+tab", "up":
			m.cursor = (m.cursor - 1 + len(m.fields)) % len(m.fields)
			m.scrollToCursor()
			return m, nil
		case "ctrl+w":
			m.addWindow("")
			m.cursor = len(m.fields) - 3
			m.scrollToCursor()
			return m, nil
		case "ctrl+d":
			if field.fieldType >= FieldWindowName {
				m.removeWindow(field.windowIdx)
				m.scrollToCursor()
			}
			return m, nil
		case "ctrl+n":
			if field.fieldType >= FieldWindowName {
				m.addPane(field.windowIdx)
				m.scrollToCursor()
			}
			return m, nil
		case "ctrl+x":
			if field.fieldType == FieldPaneCmd {
				m.removePane(m.cursor)
				m.scrollToCursor()
			}
			return m, nil
		case "enter":
			if m.isComplete() && m.validate() {
				m.Done = true
				return m, nil
			}

			if m.cursor < len(m.fields)-1 {
				m.cursor++
				m.scrollToCursor()
			} else {
				if m.validate() {
					m.Done = true
				}
			}
			return m, nil
		case "h", "l":
			if field.isChoice {
				if key == "h" {
					field.choiceIdx = (field.choiceIdx - 1 + len(field.choices)) % len(field.choices)
				} else {
					field.choiceIdx = (field.choiceIdx + 1) % len(field.choices)
				}
				return m, nil
			}
		case " ", "space":
			if !field.isChoice {
				field.value += " "
				if field.fieldType == FieldProjectDir {
					m.updateSuggestions()
				}
			}
			return m, nil
		case "backspace":
			if !field.isChoice && len(field.value) > 0 {
				field.value = field.value[:len(field.value)-1]
				if field.fieldType == FieldProjectDir {
					m.updateSuggestions()
				}
			}
			return m, nil
		}

		if !field.isChoice && len(key) == 1 {
			field.value += key
			if field.fieldType == FieldProjectDir {
				m.updateSuggestions()
			}
		}
	}
	return m, nil
}

func (m *FormModel) addPane(winIdx int) {
	insertIdx := -1
	paneCount := 0
	for i, f := range m.fields {
		if f.windowIdx == winIdx {
			insertIdx = i
			if f.fieldType == FieldPaneCmd {
				paneCount++
			}
		}
	}

	if insertIdx == -1 {
		return
	}

	newPane := FormField{
		fieldType: FieldPaneCmd,
		label:     fmt.Sprintf("Pane %d", paneCount+1),
		value:     "",
		windowIdx: winIdx,
	}

	m.fields = append(m.fields[:insertIdx+1], append([]FormField{newPane}, m.fields[insertIdx+1:]...)...)
	m.cursor = insertIdx + 1
}

func (m *FormModel) removePane(cursor int) {
	winIdx := m.fields[cursor].windowIdx
	paneCount := 0
	for _, f := range m.fields {
		if f.windowIdx == winIdx && f.fieldType == FieldPaneCmd {
			paneCount++
		}
	}
	if paneCount > 1 {
		m.fields = append(m.fields[:cursor], m.fields[cursor+1:]...)
		if m.cursor >= len(m.fields) {
			m.cursor = len(m.fields) - 1
		}
		pIdx := 1
		for i := range m.fields {
			if m.fields[i].windowIdx == winIdx && m.fields[i].fieldType == FieldPaneCmd {
				m.fields[i].label = fmt.Sprintf("Pane %d", pIdx)
				pIdx++
			}
		}
	}
}

func (m *FormModel) removeWindow(winIdx int) {
	totalWindows := 0
	for _, f := range m.fields {
		if f.fieldType == FieldWindowName {
			totalWindows++
		}
	}
	if totalWindows <= 1 {
		return
	}

	newFields := []FormField{}
	for _, f := range m.fields {
		if f.fieldType < FieldWindowName || f.windowIdx != winIdx {
			if f.windowIdx > winIdx {
				f.windowIdx--
			}
			newFields = append(newFields, f)
		}
	}
	m.fields = newFields
	if m.cursor >= len(m.fields) {
		m.cursor = len(m.fields) - 1
	}
}

func (m *FormModel) updateSuggestions() {
	field := &m.fields[1]
	input := field.value
	if input == "" {
		field.suggestion = ""
		return
	}

	path := input
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		if path == "~" {
			path = home
		} else if strings.HasPrefix(path, "~/") {
			path = filepath.Join(home, path[2:])
		}
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)

	if strings.HasSuffix(path, string(os.PathSeparator)) {
		dir = path
		base = ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		field.suggestion = ""
		return
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), base) {
			match := filepath.Join(dir, entry.Name())
			if strings.HasPrefix(input, "~") {
				home, _ := os.UserHomeDir()
				if strings.HasPrefix(match, home) {
					match = "~" + strings.TrimPrefix(match, home)
				}
			}

			if len(match) > len(input) {
				field.suggestion = match
				return
			}
		}
	}
	field.suggestion = ""
}

func (m *FormModel) validate() bool {
	return m.fields[0].value != "" && m.fields[1].value != ""
}

func (m *FormModel) View() tea.View {
	var s strings.Builder
	title := "New Project"
	if m.isEdit {
		title = "Edit Project"
	}
	s.WriteString(titleStyle.Width(m.width-2).Render(title) + "\n")

	// Available field height
	maxHeight := m.height - 12
	linesRendered := 0

	for i := m.offset; i < len(m.fields); i++ {
		f := m.fields[i]

		// Approximate line counts per field (header/spacing + input)
		fieldLines := 2
		if f.fieldType == FieldWindowName {
			fieldLines = 4
		}

		if linesRendered+fieldLines > maxHeight {
			s.WriteString(helpStyle.Render(fmt.Sprintf("\n... and %d more fields", len(m.fields)-i)) + "\n")
			break
		}

		// Indentation for hierarchy
		indentWidth := 0
		if f.fieldType == FieldWindowName || f.fieldType == FieldWindowLayout {
			indentWidth = 2
		} else if f.fieldType == FieldPaneCmd {
			indentWidth = 4
		}

		if f.fieldType == FieldWindowName {
			s.WriteString("\n" + strings.Repeat(" ", indentWidth) + labelStyle.Render(fmt.Sprintf(" Window #%d ", f.windowIdx+1)) + "\n")
			linesRendered += 2
		}

		labelStr := f.label
		labelStyle := lipgloss.NewStyle().Width(20)
		if i == m.cursor {
			labelStr = "● " + f.label
			labelStyle = lipgloss.NewStyle().Foreground(Color.Primary).Bold(true)
		} else {
			labelStr = "○ " + f.label
			labelStyle = lipgloss.NewStyle().Foreground(Color.Secondary)
		}
		labelView := labelStyle.Render(labelStr)

		var valueDisplay string
		if f.isChoice {
			var choiceStrings []string
			for idx, c := range f.choices {
				if idx == f.choiceIdx {
					choiceStrings = append(choiceStrings, focusedBorderStyle.Padding(0, 1).Render(c))
				} else {
					choiceStrings = append(choiceStrings, borderStyle.Padding(0, 1).Render(c))
				}
			}
			valueDisplay = lipgloss.JoinHorizontal(lipgloss.Center, choiceStrings...)
		} else {
			valueDisplay = f.value
			if i == m.cursor {
				valueDisplay += lipgloss.NewStyle().Foreground(Color.Primary).Blink(true).Render("|")
				if f.suggestion != "" {
					ghostText := strings.TrimPrefix(f.suggestion, f.value)
					valueDisplay += helpStyle.Render(ghostText)
				}
			}

			style := borderStyle
			if i == m.cursor {
				style = focusedBorderStyle
			}

			valueDisplay = style.Width(m.width - 20 - indentWidth - 6).Render(valueDisplay)
		}

		row := lipgloss.JoinHorizontal(lipgloss.Center, labelView, valueDisplay)
		s.WriteString(lipgloss.NewStyle().MarginLeft(indentWidth).Render(row) + "\n")
		linesRendered += 2
	}

	helpLine1 := "[tab] Move • [ctrl+w] Add Window • [ctrl+d] Del Window"
	helpLine2 := "[ctrl+n] Add Pane • [ctrl+x] Del Pane • [enter] Submit • [esc] Cancel"

	s.WriteString("\n" + statusStyle.Render(helpLine1) + "\n")
	s.WriteString(statusStyle.Render(helpLine2))

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}

func (m *FormModel) GetProjectData() (string, string, []anvil.Window) {
	name := m.fields[0].value
	dir := m.fields[1].value

	windowsMap := make(map[int]*anvil.Window)
	windowOrder := []int{}

	for _, f := range m.fields {
		if f.fieldType < FieldWindowName {
			continue
		}

		if _, ok := windowsMap[f.windowIdx]; !ok {
			windowsMap[f.windowIdx] = &anvil.Window{}
			windowOrder = append(windowOrder, f.windowIdx)
		}

		win := windowsMap[f.windowIdx]
		switch f.fieldType {
		case FieldWindowName:
			win.Name = f.value
		case FieldWindowLayout:
			win.Layout = f.choices[f.choiceIdx]
		case FieldPaneCmd:
			if f.value != "" {
				win.Panes = append(win.Panes, anvil.Pane{Cmd: f.value})
			}
		}
	}

	resWindows := []anvil.Window{}
	for _, idx := range windowOrder {
		resWindows = append(resWindows, *windowsMap[idx])
	}

	return name, dir, resWindows
}
