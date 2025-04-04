package internal

import (
	"bufio"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func remove(s []string, target string) []string {
	for i, v := range s {
		if v == target {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

var pointer = lipgloss.NewStyle().Foreground(lipgloss.Color("#eb6f92"))
var blink = lipgloss.NewStyle().Blink(true)
var ModeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#eb6f92")).Bold(true)
var help = lipgloss.NewStyle().Foreground(lipgloss.Color("#908caa"))

const ERROR_CODE = -1
const OK_CODE = 0

func StatusStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#ebbcba")).Width(width).Bold(true)
}

func splitByNewlines(str string) []string {
	reader := strings.NewReader(str)
	scanner := bufio.NewScanner(reader)

	var status []string

	for scanner.Scan() {
		line := scanner.Text()
		status = append(status, line)
	}

	if err := scanner.Err(); err != nil {
		// fmt.Println("Error reading string:", err)
	}

	return status
}
