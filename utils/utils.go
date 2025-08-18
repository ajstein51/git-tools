package utils

import "os/exec"

func IsInsideGitRepository() bool {
	checkCmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := checkCmd.Run(); err != nil {
		return false
	}
	return true
}