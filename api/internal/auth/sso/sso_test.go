package sso

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestRefreshToken_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("grant_type") != "refresh_token" {
			t.Fatalf("expected grant_type refresh_token, got %s", r.FormValue("grant_type"))
		}
		if r.FormValue("refresh_token") != "test-refresh-token" {
			t.Fatalf("unexpected refresh_token: %s", r.FormValue("refresh_token"))
		}
		resp := map[string]interface{}{
			"access_token":       "new-access-token-with-acr",
			"refresh_token":      "new-refresh-token",
			"expires_in":         300,
			"refresh_expires_in": 1800,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	adapter := &OIDCAdapter{
		oauth2Cfg: &oauth2.Config{
			Endpoint: oauth2.Endpoint{TokenURL: ts.URL},
		},
		clientID: "test-client",
	}

	result, err := adapter.RefreshToken(context.Background(), "test-refresh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "new-access-token-with-acr" {
		t.Errorf("access_token = %q, want %q", result.AccessToken, "new-access-token-with-acr")
	}
	if result.RefreshToken != "new-refresh-token" {
		t.Errorf("refresh_token = %q, want %q", result.RefreshToken, "new-refresh-token")
	}
	if result.ExpiresIn != 300 {
		t.Errorf("expires_in = %d, want 300", result.ExpiresIn)
	}
	if result.RefreshExpiresIn != 1800 {
		t.Errorf("refresh_expires_in = %d, want 1800", result.RefreshExpiresIn)
	}
}

func TestRefreshToken_InvalidGrant(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "invalid_grant", "error_description": "Token is not active"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	adapter := &OIDCAdapter{
		oauth2Cfg: &oauth2.Config{
			Endpoint: oauth2.Endpoint{TokenURL: ts.URL},
		},
		clientID: "test-client",
	}

	_, err := adapter.RefreshToken(context.Background(), "expired-token")
	if err == nil {
		t.Fatal("expected error for invalid_grant, got nil")
	}
	if !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("error should contain 'invalid_grant', got: %v", err)
	}
}

func TestRefreshToken_DerivesACRAndRequires2FA(t *testing.T) {
	// Build a minimal JWT-like access token with known claims in the payload
	accessToken := buildTestAccessToken(AccessTokenClaims{
		ACR: "mobo-2fa",
		ResourceAccess: map[string]clientRolesRaw{
			"test-client": {Roles: []string{"access", "otp_required"}},
		},
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"access_token":       accessToken,
			"refresh_token":      "rotated-refresh",
			"expires_in":         600,
			"refresh_expires_in": 3600,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	adapter := &OIDCAdapter{
		oauth2Cfg: &oauth2.Config{
			Endpoint: oauth2.Endpoint{TokenURL: ts.URL},
		},
		clientID: "test-client",
	}

	result, err := adapter.RefreshToken(context.Background(), "valid-refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ACR != "mobo-2fa" {
		t.Errorf("ACR = %q, want mobo-2fa", result.ACR)
	}
	if !result.Requires2FA {
		t.Error("Requires2FA should be true when otp_required role is present")
	}
}

func TestRefreshToken_MissingAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"refresh_token":      "only-refresh",
			"expires_in":         300,
			"refresh_expires_in": 1800,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	adapter := &OIDCAdapter{
		oauth2Cfg: &oauth2.Config{
			Endpoint: oauth2.Endpoint{TokenURL: ts.URL},
		},
		clientID: "test-client",
	}

	_, err := adapter.RefreshToken(context.Background(), "valid-refresh")
	if err == nil {
		t.Fatal("expected error for missing access_token")
	}
	if !strings.Contains(err.Error(), "no access_token") {
		t.Errorf("error should mention missing access_token, got: %v", err)
	}
}

// buildTestAccessToken creates a fake JWT with header.payload.signature format
// containing the given claims in the payload for parsing by RefreshToken.
func buildTestAccessToken(claims AccessTokenClaims) string {
	payload, _ := json.Marshal(claims)
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return "header." + encoded + ".signature"
}
