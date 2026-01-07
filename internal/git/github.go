package git

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	client *github.Client
	token  string
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(token string) *GitHubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GitHubClient{
		client: github.NewClient(tc),
		token:  token,
	}
}

// GetUserRepositories lists repositories accessible to the authenticated user
func (c *GitHubClient) GetUserRepositories(ctx context.Context) ([]*Repository, error) {
	opt := &github.RepositoryListOptions{
		Type:        "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*Repository
	for {
		repos, resp, err := c.client.Repositories.List(ctx, "", opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, repo := range repos {
			allRepos = append(allRepos, &Repository{
				ID:          repo.GetID(),
				Name:        repo.GetName(),
				FullName:    repo.GetFullName(),
				Owner:       repo.GetOwner().GetLogin(),
				Private:     repo.GetPrivate(),
				Description: repo.GetDescription(),
				URL:         repo.GetHTMLURL(),
				CloneURL:    repo.GetCloneURL(),
				DefaultBranch: repo.GetDefaultBranch(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetBranches lists branches for a repository
func (c *GitHubClient) GetBranches(ctx context.Context, owner, repo string) ([]*Branch, error) {
	branches, _, err := c.client.Repositories.ListBranches(ctx, owner, repo, &github.BranchListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var result []*Branch
	for _, branch := range branches {
		result = append(result, &Branch{
			Name:      branch.GetName(),
			Protected: branch.GetProtected(),
			CommitSHA: branch.GetCommit().GetSHA(),
		})
	}

	return result, nil
}

// GetRepositoryTree gets the directory tree for a repository
func (c *GitHubClient) GetRepositoryTree(ctx context.Context, owner, repo, branch, path string) ([]*TreeEntry, error) {
	ref := branch
	if ref == "" {
		// Get default branch
		repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository: %w", err)
		}
		ref = repository.GetDefaultBranch()
	}

	tree, _, err := c.client.Git.GetTree(ctx, owner, repo, ref, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	var result []*TreeEntry
	for _, entry := range tree.Entries {
		// Filter by path if provided
		if path != "" && path != "/" {
			if entry.GetPath() != path && !startsWith(entry.GetPath(), path+"/") {
				continue
			}
		}

		result = append(result, &TreeEntry{
			Path: entry.GetPath(),
			Type: entry.GetType(), // blob, tree
			Size: int64(entry.GetSize()),
			SHA:  entry.GetSHA(),
			URL:  entry.GetURL(),
		})
	}

	return result, nil
}

// CreateWebhook creates a webhook for a repository
func (c *GitHubClient) CreateWebhook(ctx context.Context, owner, repo string, config *WebhookConfig) (*Webhook, error) {
	contentType := "json"
	hookConfig := github.HookConfig{
		URL:         github.String(config.URL),
		ContentType: github.String(contentType),
		Secret:      github.String(config.Secret),
	}

	hook := &github.Hook{
		Name:   github.String("web"),
		Config: &hookConfig,
		Events: []string{"push"},
		Active: github.Bool(true),
	}

	createdHook, _, err := c.client.Repositories.CreateHook(ctx, owner, repo, hook)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return &Webhook{
		ID:     createdHook.GetID(),
		URL:    createdHook.GetURL(),
		Events: createdHook.Events,
		Active: createdHook.GetActive(),
	}, nil
}

// DeleteWebhook deletes a webhook
func (c *GitHubClient) DeleteWebhook(ctx context.Context, owner, repo string, hookID int64) error {
	_, err := c.client.Repositories.DeleteHook(ctx, owner, repo, hookID)
	return err
}

// Helper types
type Repository struct {
	ID            int64
	Name          string
	FullName      string
	Owner         string
	Private       bool
	Description   string
	URL           string
	CloneURL      string
	DefaultBranch string
}

type Branch struct {
	Name      string
	Protected bool
	CommitSHA string
}

type TreeEntry struct {
	Path string
	Type string // blob, tree
	Size int64
	SHA  string
	URL  string
}

type WebhookConfig struct {
	URL    string
	Secret string
}

type Webhook struct {
	ID     int64
	URL    string
	Events []string
	Active bool
}

// Helper functions
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

