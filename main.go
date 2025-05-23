package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"githuib.com/afnank19/git-tui/internal"
)

type model struct {
	height       int
	StageModel   internal.StageModel
	CommitModel  internal.CommitModel
	mode         string
	windowWidth  int
	windowHeight int
	status       string
}

func initialModel(status []string) model {
	return model{
		height:      3,
		StageModel:  internal.InitialStageModel(status),
		CommitModel: internal.InitialCommitModel(),
		mode:        "ADD",
		status:      "OK",
	}
}

func main() {
	if !internal.CheckIfGitRepo() {
		fmt.Println("No git repo found in current directory, or any parent directory")
		os.Exit(1)
	}

	cmd := exec.Command("git", "status", "--porcelain", "-uall")

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("An ERROR occurred while starting the program :(")
		fmt.Println("> This is likely due to your git repo being bad\n> Last Cmd: git status --porcelain -uall")
		fmt.Println("DEBUG: Error thrown by Cmd:", err)
		return
	}

	status := internal.GitStatusParser(string(output))

	p := tea.NewProgram(initialModel(status), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.StageModel.Init(), m.CommitModel.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case internal.CommitUpdateMsg:
		m.mode = msg.NewMode
		return m, func() tea.Msg { return internal.StageUpdateMsg{Reset: true} }

	case internal.StatusMsg:
		m.status = msg.Message

	case internal.GitPushStatus:
		m.status = string(msg)

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		case "P":
			if m.mode == "ADD" {
				m.status = "Attempting to PUSH..."
				return m, internal.RunGitPushCmd()
			}

		case "c":
			if m.mode == "ADD" {
				m.mode = "COMMIT"
			}
		case "esc":
			m.mode = "ADD"
		}
	}

	if m.mode == "ADD" {
		updated, c := m.StageModel.Update(msg)
		if sm, ok := updated.(internal.StageModel); ok {
			m.StageModel = sm
		}

		cmd = tea.Batch(cmd, c)
	} else if m.mode == "COMMIT" {
		updated, c := m.CommitModel.Update(msg)
		if sm, ok := updated.(internal.CommitModel); ok {
			m.CommitModel = sm
		}

		cmd = tea.Batch(cmd, c)
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, cmd
}

func (m model) View() string {
	statusStyle := internal.StatusStyle(m.windowWidth)
	var view string

	if m.mode == "ADD" {
		view = m.StageModel.View()
	} else if m.mode == "COMMIT" {
		view = m.CommitModel.View()
	}

	mode := " | " + internal.ModeLabel.Render("MODE:") + statusStyle.Render(" "+m.mode)

	view = view + internal.ModeLabel.Render("\nSTATUS: ") + m.status + mode
	view = lipgloss.PlaceVertical(m.windowHeight, lipgloss.Top, view)

	return view
}
