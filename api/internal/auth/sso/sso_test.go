package sso

import (
	"os"
	"testing"
	"time"

	"github.com/sed-evaluacion-desempeno/api/internal/auth"
)

func TestTransactionStore_StoreAndConsume(t *testing.T) {
	store := NewTransactionStore(5 * time.Minute)

	tx := &OIDCTransaction{State: "s1", Nonce: "n1", CodeVerifier: "cv1"}
	store.Store("s1", tx)

	got, ok := store.Consume("s1")
	if !ok {
		t.Fatal("expected to consume")
	}
	if got.Nonce != "n1" {
		t.Errorf("expected nonce n1, got %s", got.Nonce)
	}
	if got.CodeVerifier != "cv1" {
		t.Errorf("expected cv1, got %s", got.CodeVerifier)
	}

	_, ok = store.Consume("s1")
	if ok {
		t.Error("should be consumed once")
	}
}

func TestTransactionStore_Expired(t *testing.T) {
	store := NewTransactionStore(50 * time.Millisecond)

	store.Store("s1", &OIDCTransaction{State: "s1"})
	time.Sleep(100 * time.Millisecond)

	_, ok := store.Consume("s1")
	if ok {
		t.Error("expired transaction should be rejected")
	}
}

func TestTransactionStore_MultipleTabs(t *testing.T) {
	store := NewTransactionStore(5 * time.Minute)

	store.Store("tab1", &OIDCTransaction{State: "tab1"})
	store.Store("tab2", &OIDCTransaction{State: "tab2"})

	got1, ok1 := store.Consume("tab1")
	if !ok1 || got1.State != "tab1" {
		t.Error("tab1 should be consumable")
	}
	got2, ok2 := store.Consume("tab2")
	if !ok2 || got2.State != "tab2" {
		t.Error("tab2 should be consumable")
	}
}

func TestTransactionStore_InvalidState(t *testing.T) {
	store := NewTransactionStore(5 * time.Minute)
	_, ok := store.Consume("nonexistent")
	if ok {
		t.Error("nonexistent state should be rejected")
	}
}

func TestTransaction_IsStepUp(t *testing.T) {
	tx := &OIDCTransaction{RequestedACR: "mobo-2fa"}
	if !tx.IsStepUp() {
		t.Error("expected IsStepUp true")
	}
	tx2 := &OIDCTransaction{}
	if tx2.IsStepUp() {
		t.Error("expected IsStepUp false")
	}
}

func TestRequires2FA(t *testing.T) {
	claims := &AccessTokenClaims{
		ResourceAccess: map[string]clientRolesRaw{
			"sed-evaluacion": {Roles: []string{"access", "usuario"}},
		},
	}
	if Requires2FA(claims, "sed-evaluacion") {
		t.Error("should not require 2FA without otp_required")
	}

	claims2 := &AccessTokenClaims{
		ResourceAccess: map[string]clientRolesRaw{
			"sed-evaluacion": {Roles: []string{"access", "admin", "otp_required"}},
		},
	}
	if !Requires2FA(claims2, "sed-evaluacion") {
		t.Error("should require 2FA with otp_required")
	}
}

func TestHasLoA2(t *testing.T) {
	claims := &AccessTokenClaims{ACR: "mobo-2fa"}
	if !HasLoA2(claims) {
		t.Error("expected HasLoA2 true")
	}
	claims2 := &AccessTokenClaims{ACR: "1"}
	if HasLoA2(claims2) {
		t.Error("expected HasLoA2 false")
	}
	claims3 := &AccessTokenClaims{}
	if HasLoA2(claims3) {
		t.Error("expected HasLoA2 false for empty acr")
	}
}

func TestHasClientRole(t *testing.T) {
	claims := &AccessTokenClaims{
		ResourceAccess: map[string]clientRolesRaw{
			"sed-evaluacion": {Roles: []string{"access", "admin"}},
		},
	}
	if !HasClientRole(claims, "sed-evaluacion", "access") {
		t.Error("should have access role")
	}
	if !HasClientRole(claims, "sed-evaluacion", "admin") {
		t.Error("should have admin role")
	}
	if HasClientRole(claims, "sed-evaluacion", "otp_required") {
		t.Error("should not have otp_required")
	}
	if HasClientRole(claims, "other-client", "access") {
		t.Error("should not have role for other client")
	}
}

func TestRequires2FAFromLocalRole_Prod(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	if Requires2FAFromLocalRole(auth.RoleRH) {
		t.Error("should not force 2FA in prod")
	}
	if Requires2FAFromLocalRole(auth.RoleDirector) {
		t.Error("should not force 2FA in prod")
	}
	if Requires2FAFromLocalRole(auth.RoleColaborador) {
		t.Error("should not force 2FA in prod for colaborador")
	}
}

func TestRequires2FAFromLocalRole_Dev(t *testing.T) {
	os.Setenv("APP_ENV", "development")
	defer os.Unsetenv("APP_ENV")

	if !Requires2FAFromLocalRole(auth.RoleRH) {
		t.Error("should force 2FA for RH in dev")
	}
	if !Requires2FAFromLocalRole(auth.RoleDirector) {
		t.Error("should force 2FA for Director in dev")
	}
	if Requires2FAFromLocalRole(auth.RoleColaborador) {
		t.Error("should not force 2FA for colaborador in dev")
	}
}

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	if verifier == "" {
		t.Error("verifier should not be empty")
	}
	if challenge == "" {
		t.Error("challenge should not be empty")
	}
	if verifier == challenge {
		t.Error("verifier and challenge should differ")
	}
}
