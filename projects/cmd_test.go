package projects

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetLastProjectNumber(t *testing.T) {
	t.Run("Org project is newer", func(t *testing.T) {
		mockResponse := &struct {
			Organization struct {
				ProjectsV2 struct {
					Nodes []struct {
						Number    int
						CreatedAt time.Time
					}
				} `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
			} `graphql:"organization(login: $owner)"`
			Repository struct {
				ProjectsV2 struct {
					Nodes []struct {
						Number    int
						CreatedAt time.Time
					}
				} `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}{}
		mockResponse.Organization.ProjectsV2.Nodes = []struct {
			Number    int
			CreatedAt time.Time
		}{{Number: 10, CreatedAt: time.Now()}}
		mockResponse.Repository.ProjectsV2.Nodes = []struct {
			Number    int
			CreatedAt time.Time
		}{{Number: 5, CreatedAt: time.Now().Add(-time.Hour)}}

		mockClient := &mockGQLClient{mockResponse: mockResponse}
		num, err := getLastProjectNumber(mockClient, "my-org", "my-repo")
		assert.NoError(t, err)
		assert.Equal(t, 10, num)
	})

	t.Run("Client returns an error", func(t *testing.T) {
		mockClient := &mockGQLClient{mockErr: fmt.Errorf("API rate limit exceeded")}
		_, err := getLastProjectNumber(mockClient, "my-org", "my-repo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API rate limit exceeded")
	})

	t.Run("No projects found", func(t *testing.T) {
		mockResponse := &struct {
			Organization struct {
				ProjectsV2 struct {
					Nodes []struct {
						Number    int
						CreatedAt time.Time
					}
				} `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
			} `graphql:"organization(login: $owner)"`
			Repository struct {
				ProjectsV2 struct {
					Nodes []struct {
						Number    int
						CreatedAt time.Time
					}
				} `graphql:"projectsV2(first: 1, orderBy: {field: CREATED_AT, direction: DESC})"`
			} `graphql:"repository(owner: $owner, name: $repo)"`
		}{} 

		mockClient := &mockGQLClient{mockResponse: mockResponse}
		_, err := getLastProjectNumber(mockClient, "my-org", "my-repo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no projects found")
	})
}