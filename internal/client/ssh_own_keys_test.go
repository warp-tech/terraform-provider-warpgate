package client

import "testing"

func TestPublicKeyBase64IsTheKeyBody(t *testing.T) {
	key := SSHClientKey{PublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample warpgate"}
	if got := key.PublicKeyBase64(); got != "AAAAC3NzaC1lZDI1NTE5AAAAIExample" {
		t.Fatalf("expected the base64 body, got %q", got)
	}

	if got := (SSHClientKey{PublicKey: "ssh-ed25519"}).PublicKeyBase64(); got != "" {
		t.Fatalf("expected an empty body for a bare algorithm, got %q", got)
	}
}
