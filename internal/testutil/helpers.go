package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/intelifox/click-deploy/internal/auth"
)

// MockAuthContext creates a context with mock authentication values
func MockAuthContext(ctx context.Context, userID, orgID string) context.Context {
	ctx = context.WithValue(ctx, auth.UserIDKey, userID)
	ctx = context.WithValue(ctx, auth.OrgIDKey, orgID)
	ctx = context.WithValue(ctx, auth.RolesKey, []string{"user"})
	return ctx
}

// MockAuthContextPreserveRouteContext updates auth context while preserving chi route context
func MockAuthContextPreserveRouteContext(req *http.Request, userID, orgID string) *http.Request {
	ctx := req.Context()
	// Preserve chi route context if it exists
	var rctx *chi.Context
	if existingRctx, ok := ctx.Value(chi.RouteCtxKey).(*chi.Context); ok {
		rctx = existingRctx
	}
	
	// Create new auth context
	ctx = MockAuthContext(ctx, userID, orgID)
	
	// Restore chi route context if it existed
	if rctx != nil {
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	}
	
	return req.WithContext(ctx)
}

// MockRequest creates an HTTP request with mock auth context
func MockRequest(t *testing.T, method, path string, body io.Reader) (*http.Request, context.Context) {
	t.Helper()

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	ctx := MockAuthContext(req.Context(), "test-user-123", "test-org-456")
	
	// Set up chi router context for URL parameters
	rctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
	
	return req, ctx
}

// MockRequestWithURLParam creates a request with URL parameters for chi router
func MockRequestWithURLParam(t *testing.T, method, path string, params map[string]string, body io.Reader) (*http.Request, context.Context) {
	t.Helper()

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set up chi router context for URL parameters
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	
	// Add auth context (default values, can be overridden)
	ctx := MockAuthContext(context.Background(), "test-user-123", "test-org-456")
	// Set chi route context BEFORE adding to request
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)
	
	return req, ctx
}

// MockRequestWithURLParamAndAuth creates a request with URL parameters and specific auth context
func MockRequestWithURLParamAndAuth(t *testing.T, method, path string, params map[string]string, body io.Reader, userID, orgID string) (*http.Request, context.Context) {
	t.Helper()

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set up chi router context for URL parameters
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	
	// Add auth context with specified values
	ctx := MockAuthContext(context.Background(), userID, orgID)
	// Set chi route context
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)
	
	return req, ctx
}

// MockRequestJSON creates an HTTP request with JSON body and mock auth context
func MockRequestJSON(t *testing.T, method, path string, body interface{}) (*http.Request, context.Context) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	return MockRequest(t, method, path, bodyReader)
}

// MockResponseRecorder creates a response recorder for testing
func MockResponseRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// GenerateUUID generates a new UUID for testing
func GenerateUUID() uuid.UUID {
	return uuid.New()
}

// GenerateUUIDString generates a new UUID string for testing
func GenerateUUIDString() string {
	return uuid.New().String()
}
