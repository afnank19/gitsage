package internal

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func CheckIfGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--show-top-level")
	_, err := cmd.Output()
	return err == nil
}

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

// Can add a resetType if needed to use this else where
func runGitResetFile(filepath string) int {
	cmd := exec.Command("git", "reset", filepath)

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

func getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	branchName, err := cmd.Output()
	if err != nil {
		return "Error getting branch"
	}

	branchName = []byte(strings.TrimSpace(string(branchName)))

	return string(branchName)
}

func GetAllBranches() []string {
	cmd := exec.Command("git", "branch", "--format='%(refname:short)'")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	branches := splitByNewlines(string(output))

	return branches
}

func runGitCheckout(branch string) int {
	cmd := exec.Command("git", "checkout", branch)
	if err := cmd.Run(); err != nil {
		// fmt.Println("Error running git restore:", err)
		return ERROR_CODE
	}

	return OK_CODE
}

func getBranchCommits() []string {
	branch := getCurrentBranch()

	cmd := exec.Command("git", "log", branch, `--pretty=format:"%h %cr %s"`)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	commitHistory := splitByNewlines(string(out))

	return commitHistory
}

// This asynchronously runs git push status, so the UI doesnt freeze
func RunGitPushCmd() tea.Cmd {
	return func() tea.Msg {
		status := RunGitPush()
		return GitPushStatus(status)
	}
}

// This function is hard-coded for remote "origin"
// I know this is weird, but i just wanted this feature working,
// Update soon perhaps?
func RunGitPush() string {
	// Check if remote is set
	remoteCmd := exec.Command("git", "remote")

	output, err := remoteCmd.Output()
	if err != nil {
		return "Error checking remote"
	}

	remotes := splitByNewlines(string(output))

	remoteExists := false
	for i := range remotes {
		if remotes[i] == "origin" {
			// fmt.Println("foudn origin")
			remoteExists = true
		}
	}

	if !remoteExists {
		return "No remote exists"
	}

	// Find out what branch the user is on
	cmd := exec.Command("git", "branch", "--show-current")
	branchName, err := cmd.Output()
	if err != nil {
		return "Error checking current branch"
	}

	if len(branchName) == 0 {
		return "Detached HEAD"
	}

	branchName = []byte(strings.TrimSpace(string(branchName)))

	// Check if the current branch has origin set
	var originSet bool = true

	originCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	_, err = originCmd.Output()
	if err != nil {
		// no origin was set for the current branch
		originSet = false
	}

	// fmt.Println(originName)
	// If remote set up, then run git push
	if originSet {
		pushCmd := exec.Command("git", "push")
		out, err := pushCmd.CombinedOutput()
		if err != nil || strings.Contains(string(out), "Everything up-to-date") {
			// fmt.Println("err or eutd")
			return "Nothing to push or error"
		}
	} else {
		pushCmd := exec.Command("git", "push", "-u", "origin", string(branchName))
		out, err := pushCmd.CombinedOutput()
		if err != nil || strings.Contains(string(out), "Everything up-to-date") {
			// fmt.Println("err or eutd")
			return "Nothing to push or error"
		}
	} // else set upstream as well

	return "Push successful"
}
