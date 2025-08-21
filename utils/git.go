package utils

import (
	"os/exec"
	"strings"
)

func IsInsideGitRepository() bool {
	checkCmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := checkCmd.Run(); err != nil {
		return false
	}
	
	return true
}

func GetBranchNames() []string {
	if !IsInsideGitRepository() {
		return []string{}
	}

	cmd := exec.Command("git", "branch", "--all", "--format=%(refname:short)")
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "HEAD" {
			branches = append(branches, line)
		}
	}
	return branches
}
