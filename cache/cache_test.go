package cache

import (
	"fmt"
	"os"
	"testing"

	"github.com/astein-peddi/git-tooling/models"
	"github.com/stretchr/testify/assert"
)

func mockFetcher(client models.GQLClient, owner, repo, branch string, limit int) ([]models.PR, error) {
	return []models.PR{
		{Number: 101, Title: "Feat: New API"},
		{Number: 102, Title: "Fix: Bug"},
	}, nil
}

func mockFetcherWithError(client models.GQLClient, owner, repo, branch string, limit int) ([]models.PR, error) {
	return nil, fmt.Errorf("simulated API error")
}

func setupTestCache(t *testing.T, initialContent string) (PathGetter, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test_cache_*.json")
	assert.NoError(t, err)
	if initialContent != "" {
		_, err = tmpFile.WriteString(initialContent)
		assert.NoError(t, err)
	}
	tmpFile.Close()
	pathGetter := func() (string, error) {
		return tmpFile.Name(), nil
	}
	cleanup := func() {
		os.Remove(tmpFile.Name())
	}
	return pathGetter, cleanup
}

func TestFetchPRsWithCache(t *testing.T) {
	owner, repo, branch := "my-org", "my-repo", "main"

	t.Run("Cache Miss - fetches, returns data, and saves to cache", func(t *testing.T) {
		pathGetter, cleanup := setupTestCache(t, "")
		defer cleanup()
		mockHashGetter := func(branchRef string) (string, error) { return "somehash", nil }

		prs, err := FetchPRsWithCache(nil, owner, repo, branch, 0, false, mockFetcher, mockHashGetter, pathGetter)
		assert.NoError(t, err)
		assert.Len(t, prs, 2)

		cachePath, _ := pathGetter()
		content, err := os.ReadFile(cachePath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), `"number": 101`)
		assert.Contains(t, string(content), `"title": "Feat: New API"`)
	})

	t.Run("Cache Hit - returns data from cache without calling fetcher", func(t *testing.T) {
		hash := "abcdef12345"
		cacheKey := fmt.Sprintf("%s/%s:%s@%s", owner, repo, branch, hash)
		
		initialContent := fmt.Sprintf(`{
			"%s": [
				{"number": 301, "title": "Cached Title 1"},
				{"number": 302, "title": "Cached Title 2"}
			]
		}`, cacheKey)
		
		pathGetter, cleanup := setupTestCache(t, initialContent)
		defer cleanup()

		mockHashGetter := func(branchRef string) (string, error) { return hash, nil }
		failingFetcher := func(client models.GQLClient, owner, repo, branch string, limit int) ([]models.PR, error) {
			t.Fatal("fetcher was called on a cache hit")
			return nil, nil
		}

		prs, err := FetchPRsWithCache(nil, owner, repo, branch, 0, false, failingFetcher, mockHashGetter, pathGetter)
		assert.NoError(t, err)
		assert.Len(t, prs, 2)
		assert.Equal(t, 301, prs[0].Number)
		assert.Equal(t, "Cached Title 1", prs[0].Title) 
	})

	t.Run("Cache Miss due to different SHA - fetches and updates cache", func(t *testing.T) {
		oldHash := "oldhash123"
		cacheKey := fmt.Sprintf("%s/%s:%s@%s", owner, repo, branch, oldHash)
		initialContent := fmt.Sprintf(`{ "%s": [{"number": 301, "title": "Old PR"}] }`, cacheKey)
		pathGetter, cleanup := setupTestCache(t, initialContent)
		defer cleanup()

		newHash := "newhash456"
		mockHashGetter := func(branchRef string) (string, error) { return newHash, nil }

		prs, err := FetchPRsWithCache(nil, owner, repo, branch, 0, false, mockFetcher, mockHashGetter, pathGetter)
		assert.NoError(t, err)
		assert.Len(t, prs, 2) 

		cachePath, _ := pathGetter()
		content, err := os.ReadFile(cachePath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), newHash)
		assert.NotContains(t, string(content), oldHash)
	})

	t.Run("Fetcher returns an error", func(t *testing.T) {
		pathGetter, cleanup := setupTestCache(t, "")
		defer cleanup()
		mockHashGetter := func(branchRef string) (string, error) { return "somehash", nil }

		_, err := FetchPRsWithCache(nil, owner, repo, branch, 0, false, mockFetcherWithError, mockHashGetter, pathGetter)
		assert.Error(t, err)
		assert.EqualError(t, err, "simulated API error")
	})
}