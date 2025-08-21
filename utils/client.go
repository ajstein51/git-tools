package utils

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

func GetGhGraphQLClient() (*api.GraphQLClient, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w. Please verify GitHub CLI is installed and run `gh auth login`", err)
	}

	return client, nil
}

func TestAuth() {
	client, err := GetGhGraphQLClient()
	if err != nil {
		log.Fatalf("Auth failed: %v", err)
	}

	var query struct {
		Viewer struct {
			Login string
		}
	}

	err = client.Query("ViewerLogin", &query, nil)
	if err != nil {
		log.Fatalf("API request failed: %v", err)
	}

	fmt.Printf("✅ Authenticated as %s\n\n", query.Viewer.Login)
}

func GetGhUsernameGraphQL() (string, error) {
	client, err := GetGhGraphQLClient()
	if err != nil {
		return "", err
	}

	var query struct {
		Viewer struct {
			Login string
		}
	}

	err = client.Query("ViewerLogin", &query, nil)
	if err != nil {
		return "", err
	}

	return query.Viewer.Login, nil
}

func GetRepoOwnerAndName() (string, string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git remote: %w", err)
	}

	rawURL := strings.TrimSpace(string(out))

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