package utils

import (
	"fmt"
	"net/url"
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

func DoesBranchExist(branch string, localOnly bool) bool {
	if !IsInsideGitRepository() {
		return false
	}

	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err := cmd.Run(); err == nil {
		return true
	}

	if localOnly {
		return false
	}

	exists, err := DoesGhBranchExistGraphQL(branch)
	if err != nil {
		return false
	}

	return exists
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

func GetRepoOwnerAndName() (string, string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git remote: %w", err)
	}

	rawURL := strings.TrimSpace(string(out))

	return parseGitRemoteURL(rawURL)
}

func parseGitRemoteURL(rawURL string) (string, string, error) {
	if strings.HasPrefix(rawURL, "git@") {
		parts := strings.SplitN(rawURL, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("failed to parse SSH URL: %s", rawURL)
		}

		path := strings.TrimSuffix(parts[1], ".git")
		segments := strings.SplitN(path, "/", 2)
		if len(segments) != 2 {
			return "", "", fmt.Errorf("failed to parse SSH path: %s", path)
		}

		return segments[0], segments[1], nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	segments := strings.SplitN(path, "/", 2)
	if len(segments) != 2 {
		return "", "", fmt.Errorf("failed to parse path: %s", u.Path)
	}

	return segments[0], segments[1], nil
}
