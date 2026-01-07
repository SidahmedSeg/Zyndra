package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// CloneOptions specifies options for cloning a repository
type CloneOptions struct {
	URL      string
	Branch   string
	Commit   string // Optional: checkout specific commit
	Token    string // OAuth token for private repos
	Provider string // github, gitlab
}

// CloneResult contains information about the cloned repository
type CloneResult struct {
	Path       string
	CommitSHA  string
	Branch     string
	ClonedAt   time.Time
}

// CloneRepository clones a git repository to a temporary directory
func CloneRepository(ctx context.Context, opts CloneOptions, baseDir string) (*CloneResult, error) {
	// Create a unique directory for this clone
	dirName := fmt.Sprintf("repo-%d", time.Now().UnixNano())
	clonePath := filepath.Join(baseDir, dirName)

	if err := os.MkdirAll(clonePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create clone directory: %w", err)
	}

	// Build auth if token provided
	var auth *http.BasicAuth
	if opts.Token != "" {
		// For GitHub, use token as username (password can be empty)
		// For GitLab, use oauth2 token
		username := "oauth2"
		if opts.Provider == "github" {
			username = opts.Token
		}
		auth = &http.BasicAuth{
			Username: username,
			Password: opts.Token,
		}
	}

	// Clone options
	cloneOpts := &git.CloneOptions{
		URL:      opts.URL,
		Progress: os.Stdout,
		Auth:     auth,
	}

	// Set branch if specified
	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
		cloneOpts.SingleBranch = true
	}

	// Clone the repository
	repo, err := git.PlainCloneContext(ctx, clonePath, false, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitSHA := ref.Hash().String()
	branch := ref.Name().Short()

	// If specific commit requested, checkout it
	if opts.Commit != "" {
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("failed to get worktree: %w", err)
		}

		commitHash := plumbing.NewHash(opts.Commit)
		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: commitHash,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to checkout commit: %w", err)
		}

		commitSHA = opts.Commit
	}

	return &CloneResult{
		Path:      clonePath,
		CommitSHA: commitSHA,
		Branch:    branch,
		ClonedAt:  time.Now(),
	}, nil
}

// CleanupRepository removes a cloned repository directory
func CleanupRepository(path string) error {
	return os.RemoveAll(path)
}

