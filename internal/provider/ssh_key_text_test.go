package provider

import "testing"

func TestSuppressSSHKeyKindDiff(t *testing.T) {
	// Warpgate takes the generation enum and returns the algorithm name.
	for _, c := range []struct {
		old, new string
		suppress bool
	}{
		{"ssh-ed25519", "Ed25519", true},
		{"ssh-rsa", "Rsa", true},
		{"ssh-ed25519", "Rsa", false},
		{"", "Ed25519", false},
	} {
		if got := suppressSSHKeyKindDiff("kind", c.old, c.new, nil); got != c.suppress {
			t.Errorf("suppressSSHKeyKindDiff(%q, %q) = %v, want %v", c.old, c.new, got, c.suppress)
		}
	}
}

func TestSuppressSSHPublicKeyDiff(t *testing.T) {
	const body = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ3bKQmMpPLVrTEQJ1bGnqLXQ1NfFHSy7lI6BoY7DDBM"

	for _, c := range []struct {
		old, new string
		suppress bool
	}{
		{body, body + " user@host", true},
		{body, body, true},
		{body, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOther user@host", false},
		{"", body, false},
	} {
		if got := suppressSSHPublicKeyDiff("public_key", c.old, c.new, nil); got != c.suppress {
			t.Errorf("suppressSSHPublicKeyDiff(%q, %q) = %v, want %v", c.old, c.new, got, c.suppress)
		}
	}
}
