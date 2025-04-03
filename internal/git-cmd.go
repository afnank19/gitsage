package internal

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func runGitStatusUAll() []string {
	cmd := exec.Command("git", "status", "--porcelain", "-uall")

	output, err := cmd.Output()
	if err != nil {
		// fmt.Println("Error:", err)
		return []string{}
	}

	status := GitStatusParser(string(output))

	return status
}

func GitStatusParser(str string) []string {
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

func runGitAdd(filepath string) int {
	cmd := exec.Command("git", "add", filepath)

	if err := cmd.Run(); err != nil {
		// fmt.Println("Error running git add:", err)
		return ERROR_CODE
	}

	return OK_CODE
}

func runGitRestoreStagedFile(filepath string) int {
	cmd := exec.Command("git", "restore", "--staged", filepath)

	if err := cmd.Run(); err != nil {
		// fmt.Println("Error running git restore:", err)
		return ERROR_CODE
	}

	// fmt.Println("Restored " + filepath)
	return OK_CODE
}

func runGitStatus(filepath string) string {
	cmd := exec.Command("git", "status", "--porcelain", filepath)

	output, err := cmd.Output()
	if err != nil {
		// fmt.Println("Error:", err)
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

func checkStagedFiles(gitStatus []string) []string {
	var stagedFiles []string
	for i := range gitStatus {
		status, filepath := interpretGitStatus(gitStatus[i])

		if status == "A " || status == "M " || status == "MM" || status == "D " {
			stagedFiles = append(stagedFiles, filepath)
		}
	}

	return stagedFiles
}

func runGitCommit(message string) int {
	cmd := exec.Command("git", "commit", "-m", message)

	if err := cmd.Run(); err != nil {
		// fmt.Println("Error running git comm:", err)
		return ERROR_CODE
	}

	return OK_CODE
}
