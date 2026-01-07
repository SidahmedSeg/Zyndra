package auth

import (
	"context"
	"testing"
)

func TestGetUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user123")
	
	userID := GetUserID(ctx)
	if userID != "user123" {
		t.Errorf("Expected user123, got %s", userID)
	}
	
	// Test with empty context
	emptyCtx := context.Background()
	userID = GetUserID(emptyCtx)
	if userID != "" {
		t.Errorf("Expected empty string, got %s", userID)
	}
}

func TestGetOrgID(t *testing.T) {
	ctx := context.WithValue(context.Background(), OrgIDKey, "org456")
	
	orgID := GetOrgID(ctx)
	if orgID != "org456" {
		t.Errorf("Expected org456, got %s", orgID)
	}
}

func TestGetRoles(t *testing.T) {
	roles := []string{"admin", "developer"}
	ctx := context.WithValue(context.Background(), RolesKey, roles)
	
	ctxRoles := GetRoles(ctx)
	if len(ctxRoles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(ctxRoles))
	}
}

