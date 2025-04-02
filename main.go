package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	choices     []string // items on the to-do list
	stagedItems []string
	cursor      int // which to-do list item our cursor is pointing at
	height      int
	offset      int
	selected    map[int]struct{} // which to-do items are selected
	warning     string
}

func initialModel(status []string) model {
	return model{
		choices:     status,
		selected:    make(map[int]struct{}),
		warning:     "",
		height:      3,
		offset:      0,
		stagedItems: []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt", "file6.txt"},
	}
}

func main() {
	cmd := exec.Command("git", "status", "--porcelain", "-uall")

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(output))

	status := gitStatusParser(string(output))

	p := tea.NewProgram(initialModel(status), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.height = msg.Height - 5

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset--
				}
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.height {
					m.offset++
				}
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			status, filepath := interpretGitStatus(m.choices[m.cursor])

			if status == "A " || status == "M " || status == "MM" || status == "D " {
				runGitRestoreStagedFile(filepath)
				updatedStatus := runGitStatus(filepath)
				m.choices[m.cursor] = updatedStatus[:len(updatedStatus)-1]
			}

			if status == "??" || status == " M" || status == " D" {
				runGitAdd(filepath)
				updatedStatus := runGitStatus(filepath)
				m.choices[m.cursor] = updatedStatus[:len(updatedStatus)-1]
			}
			//comment for test
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "Stage/Unstage files\n\n"

	end := m.offset + m.height
	if end > len(m.choices) {
		end = len(m.choices)
	}

	for i := m.offset; i < end; i++ {
		cursor := " " // no cursor by default
		if i == m.cursor {
			cursor = ">" // cursor indicator
		}
		s += fmt.Sprintf("%s %s\n", cursor, m.choices[i])
	}

	stagedItemView := "Staged Items\n\n"

	for i := range m.stagedItems {
		stagedItemView += m.stagedItems[i] + "\n"
	}

	layout := lipgloss.JoinHorizontal(lipgloss.Left, s, stagedItemView)

	return layout
}

func gitStatusParser(str string) []string {
	reader := strings.NewReader(str)
	scanner := bufio.NewScanner(reader)

	var status []string

	for scanner.Scan() {
		line := scanner.Text()
		status = append(status, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading string:", err)
	}

	return status
}

func runGitAdd(filepath string) {
	cmd := exec.Command("git", "add", filepath)

	if err := cmd.Run(); err != nil {
		fmt.Println("Error running git add:", err)
		return
	}
}

func runGitRestoreStagedFile(filepath string) {
	cmd := exec.Command("git", "restore", "--staged", filepath)

	if err := cmd.Run(); err != nil {
		fmt.Println("Error running git restore:", err)
		return
	}

	// fmt.Println("Restored " + filepath)
}

func runGitStatus(filepath string) string {
	cmd := exec.Command("git", "status", "--porcelain", filepath)

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	// status := gitStatusParser(string(output))

	return string(output)
}

func interpretGitStatus(cmdOutputStr string) (status string, filepath string) {
	status = cmdOutputStr[0:2]
	filepath = cmdOutputStr[3:]

	return status, filepath
}
