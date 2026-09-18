package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// credentialPolicyProtocols are the keys of a credential policy, one per
// protocol Warpgate can require credentials for.
var credentialPolicyProtocols = []string{"http", "ssh", "mysql", "postgres", "kubernetes", "vnc", "rdp"}

var validCredentialKinds = map[string]bool{
	"Password":        true,
	"PublicKey":       true,
	"Totp":            true,
	"Sso":             true,
	"WebUserApproval": true,
}

// credentialPolicyLists maps each protocol key onto its slot in the policy, so
// expand and flatten stay in step with the struct without spelling out every
// protocol twice.
func credentialPolicyLists(policy *client.UserRequireCredentialsPolicy) map[string]*[]client.CredentialKind {
	return map[string]*[]client.CredentialKind{
		"http":       &policy.HTTP,
		"ssh":        &policy.SSH,
		"mysql":      &policy.MySQL,
		"postgres":   &policy.Postgres,
		"kubernetes": &policy.Kubernetes,
		"vnc":        &policy.Vnc,
		"rdp":        &policy.Rdp,
	}
}

// credentialPolicySchema is a single-item block with one list of credential
// kinds per protocol.
func credentialPolicySchema(description string) *schema.Schema {
	fields := make(map[string]*schema.Schema, len(credentialPolicyProtocols))
	for _, protocol := range credentialPolicyProtocols {
		fields[protocol] = &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		}
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		// Computed as well: Warpgate always holds a policy here, seeding a new
		// user's from the instance default, so an absent block means "keep what
		// is stored" rather than "clear it".
		Computed:    true,
		MaxItems:    1,
		Description: description,
		Elem:        &schema.Resource{Schema: fields},
	}
}

// expandCredentialPolicy converts a Terraform schema representation of credential policy
// to the Warpgate API client structure.
func expandCredentialPolicy(policyList []any) *client.UserRequireCredentialsPolicy {
	if len(policyList) == 0 {
		return nil
	}

	policyMap, _ := policyList[0].(map[string]any)
	policy := &client.UserRequireCredentialsPolicy{}

	for protocol, list := range credentialPolicyLists(policy) {
		if v, ok := policyMap[protocol]; ok && v != nil {
			*list = expandCredentialKindList(v.([]any))
		}
	}

	return policy
}

// expandCredentialKindList converts a list of credential kinds from Terraform schema format
// to the Warpgate API client format.
func expandCredentialKindList(list []any) []client.CredentialKind {
	if len(list) == 0 {
		return nil
	}

	result := make([]client.CredentialKind, len(list))
	for i, v := range list {
		result[i] = client.CredentialKind(v.(string))
	}
	return result
}

// flattenCredentialPolicy converts a Warpgate API credential policy structure
// to the Terraform schema representation.
func flattenCredentialPolicy(policy *client.UserRequireCredentialsPolicy) []any {
	if policy == nil {
		return nil
	}

	result := make(map[string]any)

	for protocol, list := range credentialPolicyLists(policy) {
		if *list != nil {
			result[protocol] = flattenCredentialKindList(*list)
		}
	}

	return []any{result}
}

// flattenCredentialKindList converts a list of credential kinds from Warpgate API format
// to the Terraform schema format.
func flattenCredentialKindList(list []client.CredentialKind) []any {
	if len(list) == 0 {
		return nil
	}

	result := make([]any, len(list))
	for i, v := range list {
		result[i] = string(v)
	}
	return result
}

// validateCredentialPolicy rejects unknown protocol keys and credential kinds
// at plan time, where the message can name the offending entry.
func validateCredentialPolicy(attribute string, raw any) error {
	credPolicies, ok := raw.([]any)
	if !ok || len(credPolicies) == 0 {
		return nil
	}

	// The block is Optional+Computed, so a user who never wrote one still gets a
	// one-element list here - holding nothing until the policy has been read
	// back from Warpgate. There is no user input to check in that case.
	if credPolicies[0] == nil {
		return nil
	}

	policy, ok := credPolicies[0].(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be a map", attribute)
	}

	for key, val := range policy {
		if _, known := credentialPolicyLists(&client.UserRequireCredentialsPolicy{})[key]; !known {
			return fmt.Errorf("unknown credential policy key: %s", key)
		}

		valueList, ok := val.([]any)
		if !ok {
			return fmt.Errorf("%s.%s must be a list", attribute, key)
		}

		for i, kind := range valueList {
			kindStr, ok := kind.(string)
			if !ok || !validCredentialKinds[kindStr] {
				return fmt.Errorf("%s.%s[%d]: %s is not a valid credential kind", attribute, key, i, kindStr)
			}
		}
	}

	return nil
}
