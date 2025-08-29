package prs

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/cli/shurcooL-graphql"
)

var pullRequestRegex = regexp.MustCompile(`\(#(\d+)\)`)

func fetchPRsForBranch(client models.GQLClient, owner, repo, branch string, limit int) ([]PR, error) {
	commits, err := fetchCommitsInBranch(client, owner, repo, branch, limit)
	if err != nil {
		return nil, err
	}

	seen := make(map[int]bool)
	var prs []PR

	for _, commit := range commits {
		if prNum, ok := extractPRNumber(commit.Message); ok {
			if !seen[prNum] {
				title := strings.SplitN(commit.Message, "\n", 2)[0]
				prs = append(prs, PR{Number: prNum, Title: title})
				seen[prNum] = true
			}
		}
	}

	return prs, nil
}

func fetchCommitsInBranch(client models.GQLClient, owner, repo, branch string, limit int) ([]Commit, error) {
	var commits []Commit
	var cursor *string
	count := 0

	for {
		var query struct {
			Repository struct {
				Ref struct {
					Target struct {
						Commit struct {
							History struct {
								Edges []struct {
									Node struct {
										Oid     string
										Message string
									}
								}
								PageInfo models.PageInfo
							} `graphql:"history(first: 100, after: $after)"`
						} `graphql:"... on Commit"`
					}
				} `graphql:"ref(qualifiedName: $branch)"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}

		variables := map[string]any{
			"owner":  graphql.String(owner),
			"repo":   graphql.String(repo),
			"branch": graphql.String(branch),
			"after":  (*graphql.String)(cursor),
		}

		if err := client.Query("CommitsInBranch", &query, variables); err != nil {
			fmt.Fprintln(os.Stderr) 
			return nil, fmt.Errorf("failed to fetch commits for branch '%s': %w", branch, err)
		}

		if (query.Repository.Ref.Target.Commit.History.Edges == nil) || (len(query.Repository.Ref.Target.Commit.History.Edges) == 0 && query.Repository.Ref.Target.Commit.History.PageInfo.EndCursor == "") {
			break 
		}
		
		edges := query.Repository.Ref.Target.Commit.History.Edges
		if len(edges) == 0 {
			break
		}

		for _, edge := range edges {
			commits = append(commits, Commit{
				Oid:     edge.Node.Oid,
				Message: edge.Node.Message,
			})
			count++
			if limit > 0 && count >= limit {
				fmt.Fprintln(os.Stderr)
				return commits, nil
			}
		}

		if !query.Repository.Ref.Target.Commit.History.PageInfo.HasNextPage {
			break
		}

		cursor = &query.Repository.Ref.Target.Commit.History.PageInfo.EndCursor
	}

	fmt.Fprintln(os.Stderr)
	return commits, nil
}

func extractPRNumber(message string) (int, bool) {
	matches := pullRequestRegex.FindStringSubmatch(message)
	if len(matches) == 2 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num, true
		}
	}
	
	return 0, false
}