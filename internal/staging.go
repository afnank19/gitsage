package internal

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var cursorStyle = lipgloss.NewStyle().Bold(true)
var margin = lipgloss.NewStyle().MarginRight(0)
var border = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#6e6a86"))
var activeBorder = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#9ccfd8"))
var stagedItem = lipgloss.NewStyle().Foreground(lipgloss.Color("#c4a7e7"))

var title = lipgloss.NewStyle().Foreground(lipgloss.Color("#ebbcba")).PaddingRight(20)

type tickMsg time.Time

type list struct {
	items  []string
	cursor int
	height int
	offset int
}

type StageModel struct {
	files        list
	branches     list
	commits      list
	focus        int
	gitAddToggle bool
	currBranch   string
	termHeight   int
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
		branches: list{
			items:  GetAllBranches(),
			height: 3,
			offset: 0,
			cursor: 0,
		},
		commits: list{
			items:  getBranchCommits(),
			height: 3,
			offset: 0,
			cursor: 0,
		},
		gitAddToggle: false,
		currBranch:   getCurrentBranch(),
	}
}

func scheduledTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m StageModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return scheduledTick()
}

func (m StageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.files.items = runGitStatusUAll()
		return m, scheduledTick()

	case StageUpdateMsg:
		if msg.Reset {
			m.files.items = runGitStatusUAll()
			m.files.cursor = 0
			m.files.offset = 0
			// m.branches.items = checkbranches(m.files.items)
			m.branches.cursor = 0
			m.branches.offset = 0
			m.commits.items = getBranchCommits()
			m.commits.cursor = 0
			m.commits.offset = 0
		}

	case tea.WindowSizeMsg:
		m.files.height = min(msg.Height-7, MAX_LINE)
		m.branches.height = min(msg.Height-7, MAX_LINE)
		m.commits.height = min(msg.Height-7, MAX_LINE)
		m.termHeight = msg.Height

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
			// m.branches.items = checkbranches(m.files.items)

		case "up", "k":
			if m.focus == 0 {
				scrollListDown(&m.files)
			} else if m.focus == 1 {
				scrollListDown(&m.branches)
			} else if m.focus == 2 {
				scrollListDown(&m.commits)
			}

		case "down", "j":
			if m.focus == 0 {
				scrollListUp(&m.files)
			} else if m.focus == 1 {
				scrollListUp(&m.branches)
			} else if m.focus == 2 {
				scrollListUp(&m.commits)
			}

		case "1":
			m.focus = 0
		case "2":
			m.focus = 1
		case "3":
			m.focus = 2
		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			if m.focus == 0 && len(m.files.items) > 0 {
				status, filepath := interpretGitStatus(m.files.items[m.files.cursor])

				if status == "A " || status == "M " || status == "MM" || status == "D " {
					runGitRestoreStagedFile(filepath)
					updatedStatus := runGitStatus(filepath)
					m.files.items[m.files.cursor] = updatedStatus[:len(updatedStatus)-1]
					// m.branches.items = remove(m.branches.items, filepath)
				}

				if status == "??" || status == " M" || status == " D" || status == "AM" {
					runGitAdd(filepath)
					updatedStatus := runGitStatus(filepath)
					m.files.items[m.files.cursor] = updatedStatus[:len(updatedStatus)-1]
					// m.branches.items = append(m.branches.items, filepath)
				}
			} else if m.focus == 1 { // Change branch mode
				stagedItems := checkStagedFiles(m.files.items)

				branch := trimFirstLast(m.branches.items[m.branches.cursor])
				fmt.Print(branch)
				if len(stagedItems) == 0 { // staged items don't exist
					code := runGitCheckout(branch)
					if code == -1 {
						// cannot switch branch
						return m, func() tea.Msg { return StatusMsg{Message: "Error: Potential Conflict between branches"} }
					}
					m.currBranch = getCurrentBranch()
					m.commits.items = getBranchCommits()

				} else {
					// change app status to let user know
					return m, func() tea.Msg { return StatusMsg{Message: "There are STAGED items, UNSTAGE to change branch"} }
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
	branchListView := "(2) LOCAL BRANCHES"
	branchListView = title.Render(branchListView)
	branchListView += "\n"

	end = min(m.branches.offset+m.branches.height, len(m.branches.items))

	for i := m.branches.offset; i < end; i++ {
		cursor := " "
		currItem := m.branches.items[i]

		if m.currBranch == trimFirstLast(currItem) {
			currItem = stagedItem.Render(currItem)
		}

		if i == m.branches.cursor {
			cursor = pointer.Render(">")
			currItem = cursorStyle.Render(currItem)
		}

		branchListView += fmt.Sprintf("%s %s\n", cursor, currItem)
	}

	branchListView = margin.Render(branchListView)

	if m.focus == 1 {
		branchListView = activeBorder.Render(branchListView)
	} else {
		branchListView = border.Render(branchListView)
	}

	// Layout stuff
	layout := lipgloss.JoinHorizontal(lipgloss.Top, s, branchListView)

	// Commit History View
	commitHistView := buildCommitHistoryView(m.commits)

	if m.focus == 2 {
		commitHistView = activeBorder.Render(commitHistView)
	} else {
		commitHistView = border.Render(commitHistView)
	}

	commitHeight := lipgloss.Height(commitHistView)
	// terminalHeight := m.files.height + 3

	if commitHeight+lipgloss.Height(layout) < m.termHeight {
		layout = lipgloss.JoinVertical(lipgloss.Top, layout, commitHistView)
	}

	testStr := "[a] - toggle git add all, [c] - COMMIT mode, [P] - git push, [enter]/[space] - toggle staging, [q] - quit"
	testStr = help.Render(testStr)
	// testStr += testStr
	branch := "\nBranch: " + m.currBranch

	output := lipgloss.JoinVertical(lipgloss.Left, layout, testStr+branch)

	return output
}

func scrollListDown(list *list) {
	if list.cursor > 0 {
		list.cursor--
		if list.cursor < list.offset {
			list.offset--
		}
	}
}

func scrollListUp(list *list) {
	if list.cursor < len(list.items)-1 {
		list.cursor++
		if list.cursor >= list.offset+list.height {
			list.offset++
		}
	}
}
