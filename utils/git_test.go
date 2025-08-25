package utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		gitDir := filepath.Join(wd, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return wd, nil 
		}
		// Move up one directory
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", errors.New("not a git repository")
		}
		wd = parent
	}
}

// --- UNIT TEST ---
func TestParseGitRemoteURL(t *testing.T) {
	testCases := []struct {
		name          string
		remoteURL     string
		expectedOwner string
		expectedRepo  string
		expectError   bool
	}{
		{
			name:          "Standard HTTPS URL",
			remoteURL:     "https://github.com/owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "Standard SSH URL",
			remoteURL:     "git@github.com:owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "HTTPS URL without .git suffix",
			remoteURL:     "https://github.com/owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "GitLab SSH URL",
			remoteURL:     "git@gitlab.com:some-group/project.git",
			expectedOwner: "some-group",
			expectedRepo:  "project",
			expectError:   false,
		},
		{
			name:          "Bitbucket HTTPS URL",
			remoteURL:     "https://user@bitbucket.org/team/repository.git",
			expectedOwner: "team",
			expectedRepo:  "repository",
			expectError:   false,
		},
		{
			name:        "Invalid HTTPS URL",
			remoteURL:   "https://github.com/just-owner",
			expectError: true,
		},
		{
			name:        "Invalid SSH URL",
			remoteURL:   "git@github.com:just-owner.git",
			expectError: true,
		},
		{
			name:        "Malformed SSH URL",
			remoteURL:   "git@github.com-owner/repo.git",
			expectError: true,
		},
		{
			name:        "Completely invalid URL",
			remoteURL:   "not a url",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo, err := parseGitRemoteURL(tc.remoteURL)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOwner, owner)
				assert.Equal(t, tc.expectedRepo, repo)
			}
		})
	}
}


// --- INTEGRATION TESTS ---
func TestIsInsideGitRepository_Integration(t *testing.T) {
	_, err := getProjectRoot()
	if err != nil {
		t.Skip("skipping integration test: not in a git repository")
	}

	assert.True(t, IsInsideGitRepository(), "should return true when inside a git repository")
}

func TestGetBranchNames_Integration(t *testing.T) {
	_, err := getProjectRoot()
	if err != nil {
		t.Skip("skipping integration test: not in a git repository")
	}

	branches := GetBranchNames()
	assert.NotEmpty(t, branches, "should return a non-empty list of branches")
}

func TestGetRepoOwnerAndName_Integration(t *testing.T) {
	_, err := getProjectRoot()
	if err != nil {
		t.Skip("skipping integration test: not in a git repository")
	}

	root, _ := getProjectRoot()
	originalWd, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(originalWd)

	owner, repo, err := GetRepoOwnerAndName()
	assert.NoError(t, err, "should not return an error when a remote 'origin' exists")
	assert.NotEmpty(t, owner, "owner should not be empty")
	assert.NotEmpty(t, repo, "repo should not be empty")
}