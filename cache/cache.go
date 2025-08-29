package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/astein-peddi/git-tooling/models"
)

type prCacheData map[string][]int

type Fetcher func(client models.GQLClient, owner, repo, branch string, limit int) ([]models.PR, error)
type HashGetter func(branchRef string) (string, error)
type PathGetter func() (string, error)

func FetchPRsWithCache(client models.GQLClient, owner, repo, branch string, limit int, isLocal bool, fetcher Fetcher, hashGetter HashGetter, pathGetter PathGetter) ([]models.PR, error) {
	branchRef := branch
	if !isLocal {
		branchRef = "origin/" + branch
	}

	hash, err := hashGetter(branchRef)
	if err != nil {
		return fetcher(client, owner, repo, branch, limit)
	}

	cache, err := loadCache(pathGetter)
	if err != nil {
		return nil, fmt.Errorf("could not load cache: %w", err)
	}

	cacheKey := fmt.Sprintf("%s/%s:%s@%s", owner, repo, branch, hash)
	if prNumbers, found := cache[cacheKey]; found {
		fmt.Fprintf(os.Stderr, "Cache hit for branch '%s'. Loading %d PRs instantly.\n", branch, len(prNumbers))
		var cachedPRs []models.PR
		for _, num := range prNumbers {
			cachedPRs = append(cachedPRs, models.PR{Number: num})
		}

		return cachedPRs, nil
	}

	livePRs, err := fetcher(client, owner, repo, branch, limit)
	if err != nil {
		return nil, err
	}

	prefixToClean := fmt.Sprintf("%s/%s:%s@", owner, repo, branch)
	for k := range cache {
		if strings.HasPrefix(k, prefixToClean) {
			delete(cache, k)
		}
	}

	var prNumbers []int
	for _, pr := range livePRs {
		prNumbers = append(prNumbers, pr.Number)
	}

	cache[cacheKey] = prNumbers
	if err := saveCache(cache, pathGetter); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save cache: %v\n", err)
	}

	return livePRs, nil
}

func GetCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	toolCachePath := filepath.Join(cacheDir, "peddi-tooling")
	if err := os.MkdirAll(toolCachePath, 0755); err != nil {
		return "", err
	}

	return filepath.Join(toolCachePath, "prs_cache.json"), nil
}

func GetBranchHeadHash(branchRef string) (string, error) {
	cmd := exec.Command("git", "rev-parse", branchRef)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not get SHA for branch '%s': %w", branchRef, err)
	}

	return strings.TrimSpace(string(out)), nil
}

func loadCache(pathGetter PathGetter) (prCacheData, error) {
	path, err := pathGetter()
	if err != nil {
		return nil, err
	}

	data := make(prCacheData)
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return data, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(content, &data); err != nil {
		return make(prCacheData), nil
	}

	return data, nil
}

func saveCache(data prCacheData, pathGetter PathGetter) error {
	path, err := pathGetter()
	if err != nil {
		return err
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}
