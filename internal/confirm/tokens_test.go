// ABOUTME: Tests for confirmation token generation and validation.
// ABOUTME: Verifies token creation, validation, and expiry behavior.

package confirm

import (
	"testing"
	"time"
)

func TestTokenStore_GenerateAndValidate(t *testing.T) {
	store := NewTokenStore(5 * time.Minute)

	token := store.Generate("delete_deployment", "cf")

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	valid := store.Validate(token, "delete_deployment", "cf")
	if !valid {
		t.Error("expected token to be valid")
	}

	// Token should be consumed after validation
	valid = store.Validate(token, "delete_deployment", "cf")
	if valid {
		t.Error("expected token to be consumed after use")
	}
}

func TestTokenStore_WrongOperation(t *testing.T) {
	store := NewTokenStore(5 * time.Minute)

	token := store.Generate("delete_deployment", "cf")

	valid := store.Validate(token, "recreate", "cf")
	if valid {
		t.Error("expected token to be invalid for wrong operation")
	}
}

func TestTokenStore_WrongResource(t *testing.T) {
	store := NewTokenStore(5 * time.Minute)

	token := store.Generate("delete_deployment", "cf")

	valid := store.Validate(token, "delete_deployment", "other-deployment")
	if valid {
		t.Error("expected token to be invalid for wrong resource")
	}
}

func TestTokenStore_Expiry(t *testing.T) {
	store := NewTokenStore(50 * time.Millisecond)

	token := store.Generate("delete_deployment", "cf")

	time.Sleep(100 * time.Millisecond)

	valid := store.Validate(token, "delete_deployment", "cf")
	if valid {
		t.Error("expected token to be expired")
	}
}

func TestTokenStore_GetPendingToken(t *testing.T) {
	store := NewTokenStore(5 * time.Minute)

	token := store.Generate("delete_deployment", "cf")

	pending := store.GetPending(token)
	if pending == nil {
		t.Fatal("expected pending token info")
	}
	if pending.Operation != "delete_deployment" {
		t.Errorf("expected operation delete_deployment, got %s", pending.Operation)
	}
	if pending.Resource != "cf" {
		t.Errorf("expected resource cf, got %s", pending.Resource)
	}
}
