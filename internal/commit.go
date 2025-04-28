package internal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CommitUpdateMsg struct {
	NewMode string
}

type CommitModel struct {
	message string
}

func InitialCommitModel() CommitModel {
	return CommitModel{
		message: "",
	}
}

func (m CommitModel) Init() tea.Cmd {
	return nil
}

func (m CommitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(m.message) < 1 {
				return m, func() tea.Msg { return StatusMsg{Message: "Add a message before commiting!"} }
			}
			statusCode := runGitCommit(m.message)
			if statusCode != -1 {
				m.message = ""
				return m, func() tea.Msg { return CommitUpdateMsg{NewMode: "ADD"} }
			} else {
				return m, func() tea.Msg { return StatusMsg{Message: "Nothing to commit, stage changed first perhaps?"} }
			}

		case "backspace":
			if len(m.message) > 0 {
				m.message = m.message[:len(m.message)-1]
			}

		default:
			m.message += msg.String()
		}
	}

	return m, nil
}

func (m CommitModel) View() string {
	cursor := blink.Render("|")
	s := "Message: " + m.message + cursor

	helpText := "[esc] - ADD mode, [enter] - git commit"
	helpText = help.Render(helpText)

	out := s + "\n" + helpText

	view := lipgloss.PlaceVertical(0, lipgloss.Top, out)

	return view
}
