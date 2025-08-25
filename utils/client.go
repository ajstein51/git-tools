package utils

import (
	"fmt"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/cli/go-gh/v2/pkg/api"
)

func GetGhGraphQLClient() (models.GQLClient, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w. Please verify GitHub CLI is installed and run `gh auth login`", err)
	}

	return client, nil
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