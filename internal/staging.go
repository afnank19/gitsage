package internal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var cursorStyle = lipgloss.NewStyle().Bold(true)
var margin = lipgloss.NewStyle().MarginRight(0)
var border = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#6e6a86"))
var activeBorder = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#9ccfd8"))
var stagedItem = lipgloss.NewStyle().Foreground(lipgloss.Color("#c4a7e7"))

var title = lipgloss.NewStyle().Foreground(lipgloss.Color("#ebbcba")).PaddingRight(25)

type list struct {
	items  []string
	cursor int
	height int
	offset int
}

type StageModel struct {
	files        list
	stagedFiles  list
	focus        int
	gitAddToggle bool
}

type StageUpdateMsg struct {
	Reset bool
}

func InitialStageModel(status []string) StageModel {
	return StageModel{
		focus: 0,
		files: list{
			items:  status,
			height: 3,
			offset: 0,
			cursor: 0,
		},
		stagedFiles: list{
			items:  checkStagedFiles(status),
			height: 3,
			offset: 0,
			cursor: 0,
		},
		gitAddToggle: false,
	}
}

func (m StageModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m StageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StageUpdateMsg:
		if msg.Reset {
			m.files.items = runGitStatusUAll()
			m.files.cursor = 0
			m.files.offset = 0
			m.stagedFiles.items = checkStagedFiles(m.files.items)
			m.stagedFiles.cursor = 0
			m.stagedFiles.offset = 0
		}

	case tea.WindowSizeMsg:
		m.files.height = msg.Height - 7
		m.stagedFiles.height = msg.Height - 7

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		case "a":
			if m.gitAddToggle {
				runGitRestoreStagedFile(".")
				m.gitAddToggle = false
			} else {
				runGitAdd(".")
				m.gitAddToggle = true
			}
			m.files.items = runGitStatusUAll()
			m.stagedFiles.items = checkStagedFiles(m.files.items)

		case "up", "k":
			if m.focus == 0 {
				if m.files.cursor > 0 {
					m.files.cursor--
					if m.files.cursor < m.files.offset {
						m.files.offset--
					}
				}
			} else if m.focus == 1 {
				if m.stagedFiles.cursor > 0 {
					m.stagedFiles.cursor--
					if m.stagedFiles.cursor < m.stagedFiles.offset {
						m.stagedFiles.offset--
					}
				}
			}

		case "down", "j":
			if m.focus == 0 {
				if m.files.cursor < len(m.files.items)-1 {
					m.files.cursor++
					if m.files.cursor >= m.files.offset+m.files.height {
						m.files.offset++
					}
				}
			} else if m.focus == 1 {
				if m.stagedFiles.cursor < len(m.stagedFiles.items)-1 {
					m.stagedFiles.cursor++
					if m.stagedFiles.cursor >= m.stagedFiles.offset+m.stagedFiles.height {
						m.stagedFiles.offset++
					}
				}
			}

		case "1":
			m.focus = 0
		case "2":
			m.focus = 1
		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			if m.focus == 0 {
				status, filepath := interpretGitStatus(m.files.items[m.files.cursor])

				if status == "A " || status == "M " || status == "MM" || status == "D " {
					runGitRestoreStagedFile(filepath)
					updatedStatus := runGitStatus(filepath)
					m.files.items[m.files.cursor] = updatedStatus[:len(updatedStatus)-1]
					m.stagedFiles.items = remove(m.stagedFiles.items, filepath)
				}

				if status == "??" || status == " M" || status == " D" || status == "AM" {
					runGitAdd(filepath)
					updatedStatus := runGitStatus(filepath)
					m.files.items[m.files.cursor] = updatedStatus[:len(updatedStatus)-1]
					m.stagedFiles.items = append(m.stagedFiles.items, filepath)
				}
			}
			//comment for test
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m StageModel) View() string {

	// Stage-able Item List View logic
	s := "(1) STAGE/UNSTAGE FILES"
	s = title.Render(s)
	s += "\n"

	end := min(m.files.offset+m.files.height, len(m.files.items))

	for i := m.files.offset; i < end; i++ {
		cursor := " " // no cursor by default
		currItem := m.files.items[i]

		status, _ := interpretGitStatus(currItem)
		if status == "A " || status == "M " || status == "MM" || status == "D " {
			currItem = stagedItem.Render(currItem)
		}

		if i == m.files.cursor {
			cursor = pointer.Render(">") // cursor indicator
			currItem = cursorStyle.Render(currItem)
		}

		s += fmt.Sprintf("%s %s\n", cursor, currItem)
	}

	s = margin.Render(s)
	if m.focus == 0 {
		s = activeBorder.Render(s)
	} else {
		s = border.Render(s)
	}

	// Staged Item List View logic
	stagedItemView := "(2) STAGED FILES"
	stagedItemView = title.Render(stagedItemView)
	stagedItemView += "\n"

	end = min(m.stagedFiles.offset+m.stagedFiles.height, len(m.stagedFiles.items))

	for i := m.stagedFiles.offset; i < end; i++ {
		cursor := " "
		currItem := m.stagedFiles.items[i]
		if i == m.stagedFiles.cursor {
			cursor = pointer.Render(">")
			currItem = cursorStyle.Render(currItem)
		}
		stagedItemView += fmt.Sprintf("%s %s\n", cursor, currItem)
	}

	stagedItemView = margin.Render(stagedItemView)

	if m.focus == 1 {
		stagedItemView = activeBorder.Render(stagedItemView)
	} else {
		stagedItemView = border.Render(stagedItemView)
	}

	// Layout stuff
	layout := lipgloss.JoinHorizontal(lipgloss.Top, s, stagedItemView)

	testStr := "[a] - toggle git add all, [c] - COMMIT mode, [P] - git push, [enter]/[space] - toggle staging, [q] - quit"
	testStr = help.Render(testStr)
	// testStr += testStr

	output := lipgloss.JoinVertical(lipgloss.Left, layout, testStr)

	return output
}
