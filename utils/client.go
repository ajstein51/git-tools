package utils

import (
	"fmt"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/shurcooL-graphql"
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

func DoesGhBranchExistGraphQL(branch string) (bool, error) {
	client, err := GetGhGraphQLClient()
	if err != nil {
		return false, err
	}

	owner, repo, err := GetRepoOwnerAndName()
	if err != nil {
		return false, err
	}

	var query struct {
		Repository struct {
			Ref *struct {
				ID string
			}
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{
		"owner":  graphql.String(owner),
		"repo":   graphql.String(repo),
		"branch": graphql.String("refs/heads/" + branch),
	}

	err = client.Query("BranchExists", &query, variables)
	if err != nil {
		return false, err
	}

	return query.Repository.Ref != nil, nil
}
