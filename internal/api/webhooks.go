package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/git"
	"github.com/intelifox/click-deploy/internal/store"
)

type WebhookHandler struct {
	store  *store.DB
	config *config.Config
}

func NewWebhookHandler(store *store.DB, cfg *config.Config) *WebhookHandler {
	return &WebhookHandler{
		store:  store,
		config: cfg,
	}
}

// RegisterWebhookRoutes registers webhook routes
func RegisterWebhookRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h := NewWebhookHandler(db, cfg)

	// Webhook endpoints (public, but validated via signature)
	r.Post("/webhooks/github", h.HandleGitHubWebhook)
	r.Post("/webhooks/gitlab", h.HandleGitLabWebhook)
}

// HandleGitHubWebhook handles GitHub webhook events
func (h *WebhookHandler) HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Get signature from header
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		http.Error(w, "Missing signature", http.StatusUnauthorized)
		return
	}

	// Validate signature
	if !git.ValidateGitHubWebhookSignature(h.config.WebhookSecret, payload, signature) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "Missing event type", http.StatusBadRequest)
		return
	}

	// Parse event
	event, err := git.ParseGitHubEvent(eventType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Handle ping event (webhook test)
	if event == "ping" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
		return
	}

	// Parse push event
	if event == "push" {
		var pushEvent GitHubPushEvent
		if err := json.Unmarshal(payload, &pushEvent); err != nil {
			http.Error(w, "Failed to parse payload", http.StatusBadRequest)
			return
		}

		// Find services matching this repository and trigger deployments
		if err := h.triggerDeploymentsForPush(r.Context(), pushEvent.Repository.FullName, pushEvent.Ref, pushEvent.After, pushEvent.HeadCommit.Message, pushEvent.HeadCommit.Author.Name); err != nil {
			log.Printf("Error triggering deployments: %v", err)
			// Don't fail the webhook, just log
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleGitLabWebhook handles GitLab webhook events
func (h *WebhookHandler) HandleGitLabWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Get token from header
	token := r.Header.Get("X-Gitlab-Token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// Validate token
	if !git.ValidateGitLabWebhookSignature(h.config.WebhookSecret, token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse event type
	eventType := r.Header.Get("X-Gitlab-Event")
	if eventType == "" {
		http.Error(w, "Missing event type", http.StatusBadRequest)
		return
	}

	// Parse event
	event, err := git.ParseGitLabEvent(eventType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Handle push event
	if event == "Push Hook" {
		var pushEvent GitLabPushEvent
		if err := json.Unmarshal(payload, &pushEvent); err != nil {
			http.Error(w, "Failed to parse payload", http.StatusBadRequest)
			return
		}

		// Find services matching this repository and trigger deployments
		if len(pushEvent.Commits) > 0 {
			lastCommit := pushEvent.Commits[len(pushEvent.Commits)-1]
			if err := h.triggerDeploymentsForPush(r.Context(), pushEvent.Project.PathWithNamespace, pushEvent.Ref, pushEvent.After, lastCommit.Message, lastCommit.Author.Name); err != nil {
				log.Printf("Error triggering deployments: %v", err)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GitHubPushEvent represents a GitHub push webhook event
type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"head_commit"`
}

// GitLabPushEvent represents a GitLab push webhook event
type GitLabPushEvent struct {
	Ref    string `json:"ref"`
	After  string `json:"after"`
	Project struct {
		PathWithNamespace string `json:"path_with_namespace"`
	} `json:"project"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"commits"`
}

// triggerDeploymentsForPush triggers deployments for services matching the repository
func (h *WebhookHandler) triggerDeploymentsForPush(ctx context.Context, repoFullName, ref, commitSHA, commitMessage, commitAuthor string) error {
	// Extract owner and repo name
	parts := strings.Split(repoFullName, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid repository name: %s", repoFullName)
	}
	owner := parts[0]
	repoName := strings.Join(parts[1:], "/")

	// Extract branch from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(ref, "refs/heads/")

	// TODO: Find all git sources matching this repository and branch
	// This requires a query like:
	// SELECT services.* FROM services 
	// JOIN git_sources ON services.id = git_sources.service_id 
	// WHERE git_sources.repo_owner = $1 AND git_sources.repo_name = $2 AND git_sources.branch = $3
	
	// For now, log the webhook event
	log.Printf("Webhook push event: repo=%s/%s, branch=%s, commit=%s", owner, repoName, branch, commitSHA)
	
	// TODO: Implement service lookup and deployment creation
	// When implemented, create deployment and queue build job for each matching service

	return nil
}

