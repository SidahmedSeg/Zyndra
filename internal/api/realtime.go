package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/realtime"
	"github.com/intelifox/click-deploy/internal/store"
)

type RealtimeHandler struct {
	store *store.DB
	cfg   *config.Config
}

func NewRealtimeHandler(db *store.DB, cfg *config.Config) *RealtimeHandler {
	return &RealtimeHandler{store: db, cfg: cfg}
}

func RegisterRealtimeRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h := NewRealtimeHandler(db, cfg)
	r.Get("/realtime/connect-token", h.GetConnectToken)
	r.Get("/realtime/subscription-token", h.GetSubscriptionToken)
}

type connectTokenResponse struct {
	Token string `json:"token"`
	WSURL string `json:"ws_url"`
}

func (h *RealtimeHandler) GetConnectToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if h.cfg.CentrifugoTokenHMACSecret == "" || h.cfg.CentrifugoWSURL == "" {
		http.Error(w, "Centrifugo is not configured", http.StatusNotImplemented)
		return
	}

	token, err := realtime.GenerateConnectionToken(h.cfg.CentrifugoTokenHMACSecret, userID, 30*time.Minute)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, connectTokenResponse{
		Token: token,
		WSURL: h.cfg.CentrifugoWSURL,
	})
}

type subscriptionTokenResponse struct {
	Token string `json:"token"`
}

func (h *RealtimeHandler) GetSubscriptionToken(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	userID := auth.GetUserID(r.Context())
	if orgID == "" || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if h.cfg.CentrifugoTokenHMACSecret == "" {
		http.Error(w, "Centrifugo is not configured", http.StatusNotImplemented)
		return
	}

	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Missing channel", http.StatusBadRequest)
		return
	}

	// Only allow specific channel prefixes for now.
	// - deployment:<uuid>
	// - service:<uuid>
	if strings.HasPrefix(channel, "deployment:") {
		idStr := strings.TrimPrefix(channel, "deployment:")
		deploymentID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
			return
		}

		deployment, err := h.store.GetDeployment(r.Context(), deploymentID)
		if err != nil || deployment == nil {
			http.Error(w, "Deployment not found", http.StatusNotFound)
			return
		}

		service, err := h.store.GetService(r.Context(), deployment.ServiceID)
		if err != nil || service == nil {
			http.Error(w, "Deployment not found", http.StatusNotFound)
			return
		}

		project, err := h.store.GetProject(r.Context(), service.ProjectID)
		if err != nil || project == nil || project.CasdoorOrgID != orgID {
			http.Error(w, "Deployment not found", http.StatusNotFound)
			return
		}
	} else if strings.HasPrefix(channel, "service:") {
		idStr := strings.TrimPrefix(channel, "service:")
		serviceID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid service ID", http.StatusBadRequest)
			return
		}

		service, err := h.store.GetService(r.Context(), serviceID)
		if err != nil || service == nil {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}

		project, err := h.store.GetProject(r.Context(), service.ProjectID)
		if err != nil || project == nil || project.CasdoorOrgID != orgID {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, "Unsupported channel", http.StatusBadRequest)
		return
	}

	token, err := realtime.GenerateSubscriptionToken(h.cfg.CentrifugoTokenHMACSecret, userID, channel, 30*time.Minute)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, subscriptionTokenResponse{Token: token})
}


