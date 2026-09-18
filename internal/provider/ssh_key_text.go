package provider

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// normalizeSSHKeyKind reduces both spellings of a key type to one token: a key
// is created with the generation enum ("Ed25519", "Rsa") and read back as the
// OpenSSH algorithm name ("ssh-ed25519", "ssh-rsa").
func normalizeSSHKeyKind(kind string) string {
	return strings.TrimPrefix(strings.ToLower(kind), "ssh-")
}

// normalizeSSHPublicKey drops the trailing comment. Warpgate stores only the
// algorithm and the key body, so a key configured with the usual "user@host"
// comment would otherwise never match what is read back.
func normalizeSSHPublicKey(key string) string {
	fields := strings.Fields(key)
	if len(fields) > 2 {
		fields = fields[:2]
	}

	return strings.Join(fields, " ")
}

func suppressSSHKeyKindDiff(_, old, new string, _ *schema.ResourceData) bool {
	return normalizeSSHKeyKind(old) == normalizeSSHKeyKind(new)
}

func suppressSSHPublicKeyDiff(_, old, new string, _ *schema.ResourceData) bool {
	return normalizeSSHPublicKey(old) == normalizeSSHPublicKey(new)
}
