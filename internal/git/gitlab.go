package git

import (
	"context"
	"fmt"

	"github.com/xanzy/go-gitlab"
)

type GitLabClient struct {
	client *gitlab.Client
	token  string
}

// NewGitLabClient creates a new GitLab API client
func NewGitLabClient(token, baseURL string) *GitLabClient {
	client, _ := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))

	return &GitLabClient{
		client: client,
		token:  token,
	}
}

// GetUserRepositories lists repositories accessible to the authenticated user
func (c *GitLabClient) GetUserRepositories(ctx context.Context) ([]*Repository, error) {
	opt := &gitlab.ListProjectsOptions{
		Owned:  gitlab.Bool(true),
		Simple: gitlab.Bool(false),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	var allRepos []*Repository
	for {
		projects, resp, err := c.client.Projects.ListProjects(opt, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range projects {
			// GitLab uses Visibility field instead of Public
			isPrivate := project.Visibility == gitlab.PrivateVisibility || project.Visibility == gitlab.InternalVisibility
			allRepos = append(allRepos, &Repository{
				ID:            int64(project.ID),
				Name:          project.Name,
				FullName:      project.PathWithNamespace,
				Owner:         project.Namespace.FullPath,
				Private:       isPrivate,
				Description:   project.Description,
				URL:           project.WebURL,
				CloneURL:      project.HTTPURLToRepo,
				DefaultBranch: project.DefaultBranch,
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
func (c *GitLabClient) GetBranches(ctx context.Context, owner, repo string) ([]*Branch, error) {
	projectID := fmt.Sprintf("%s/%s", owner, repo)
	branches, _, err := c.client.Branches.ListBranches(projectID, &gitlab.ListBranchesOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var result []*Branch
	for _, branch := range branches {
		result = append(result, &Branch{
			Name:      branch.Name,
			Protected: branch.Protected,
			CommitSHA: branch.Commit.ID,
		})
	}

	return result, nil
}

// GetRepositoryTree gets the directory tree for a repository
func (c *GitLabClient) GetRepositoryTree(ctx context.Context, owner, repo, branch, path string) ([]*TreeEntry, error) {
	projectID := fmt.Sprintf("%s/%s", owner, repo)
	ref := branch
	if ref == "" {
		// Get default branch
		project, _, err := c.client.Projects.GetProject(projectID, nil, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to get project: %w", err)
		}
		ref = project.DefaultBranch
	}

	opt := &gitlab.ListTreeOptions{
		Ref:  gitlab.String(ref),
		Path: gitlab.String(path),
	}

	tree, _, err := c.client.Repositories.ListTree(projectID, opt, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	var result []*TreeEntry
	for _, entry := range tree {
		result = append(result, &TreeEntry{
			Path: entry.Path,
			Type: entry.Type, // blob, tree
			Size: 0,          // GitLab doesn't provide size in tree listing
			SHA:  entry.ID,
			URL:  "",
		})
	}

	return result, nil
}

// CreateWebhook creates a webhook for a repository
func (c *GitLabClient) CreateWebhook(ctx context.Context, owner, repo string, config *WebhookConfig) (*Webhook, error) {
	projectID := fmt.Sprintf("%s/%s", owner, repo)
	hook := &gitlab.AddProjectHookOptions{
		URL:                   gitlab.String(config.URL),
		PushEvents:            gitlab.Bool(true),
		Token:                 gitlab.String(config.Secret),
		EnableSSLVerification: gitlab.Bool(true),
	}

	createdHook, _, err := c.client.Projects.AddProjectHook(projectID, hook, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	var events []string
	if createdHook.PushEvents {
		events = append(events, "push")
	}

	return &Webhook{
		ID:     int64(createdHook.ID),
		URL:    createdHook.URL,
		Events: events,
		Active: true,
	}, nil
}

// DeleteWebhook deletes a webhook
func (c *GitLabClient) DeleteWebhook(ctx context.Context, owner, repo string, hookID int64) error {
	projectID := fmt.Sprintf("%s/%s", owner, repo)
	_, err := c.client.Projects.DeleteProjectHook(projectID, int(hookID), gitlab.WithContext(ctx))
	return err
}


