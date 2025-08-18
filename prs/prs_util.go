package prs

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func ListPullRequestsBetweenBranches(branchA, branchB string, useLocal bool) ([]PR, error) {
	a := branchA
	b := branchB

	if !useLocal {
		fmt.Println("Using remote branches: origin/" + branchA + " and origin/" + branchB + "\n")

		fetchCmd := exec.Command("git", "fetch", "origin", branchA, branchB)
		if err := fetchCmd.Run(); err != nil {
			return nil, fmt.Errorf("git fetch error: %v", err)
		}

		a = "origin/" + branchA
		b = "origin/" + branchB
	}

	cmd := exec.Command("git", "log", b+".."+a, "--pretty=format:%H %s")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git log error: %v\n%s", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	prRegex := regexp.MustCompile(`\(#(\d+)\)$`)

	var prs []PR
	for _, line := range lines {
		if match := prRegex.FindStringSubmatch(line); match != nil {
			prNumber := match[1]
			title := ExtractTitle(line)
			repoURL := "https://github.com/peddinghaus/raptor/pull/"

			prs = append(prs, PR{
				Number:      prNumber,
				Title:       title,
				ShortCommit: line[:7],
				URL:         fmt.Sprintf("%s%s", repoURL, prNumber),
			})
		}
	}

	return prs, nil
}

func ExtractTitle(commit string) string {
	parts := strings.SplitN(commit, " ", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(regexp.MustCompile(`\(#\d+\)$`).ReplaceAllString(parts[1], ""))
	}
	return commit
}
