package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceSSHKeySchema(t *testing.T) {
	s := resourceSSHKey()
	if s == nil {
		t.Fatal("expected non-nil resourceSSHKey")
	}

	requiredFields := []string{"label"}
	for _, field := range requiredFields {
		f, ok := s.Schema[field]
		if !ok {
			t.Fatalf("expected field %q in resourceSSHKey schema", field)
		}
		if !f.Required {
			t.Fatalf("expected field %q to be required", field)
		}
	}

	optionalFields := []string{"kind", "secret_key", "is_default"}
	for _, field := range optionalFields {
		f, ok := s.Schema[field]
		if !ok {
			t.Fatalf("expected field %q in resourceSSHKey schema", field)
		}
		if !f.Optional {
			t.Fatalf("expected field %q to be optional", field)
		}
	}

	computedFields := []string{"id", "public_key", "public_key_base64"}
	for _, field := range computedFields {
		f, ok := s.Schema[field]
		if field == "id" {
			continue // id field is handled by terraform SDK
		}
		if !ok {
			t.Fatalf("expected field %q in resourceSSHKey schema", field)
		}
		if !f.Computed {
			t.Fatalf("expected field %q to be computed", field)
		}
	}
}

func TestDataSourceSSHKeySchema(t *testing.T) {
	s := dataSourceSSHKey()
	if s == nil {
		t.Fatal("expected non-nil dataSourceSSHKey")
	}

	for _, field := range []string{"id", "label", "kind", "public_key", "public_key_base64", "is_default"} {
		if _, ok := s.Schema[field]; !ok {
			t.Fatalf("expected field %q in dataSourceSSHKey schema", field)
		}
	}
}

func TestDataSourceSSHOwnKeysSchema(t *testing.T) {
	s := dataSourceSSHOwnKeys()
	if s == nil {
		t.Fatal("expected non-nil dataSourceSSHOwnKeys")
	}

	keysSchema, ok := s.Schema["keys"]
	if !ok {
		t.Fatal("expected 'keys' in dataSourceSSHOwnKeys schema")
	}

	elemResource, ok := keysSchema.Elem.(*schema.Resource)
	if !ok {
		t.Fatal("expected keys.Elem to be *schema.Resource")
	}

	for _, field := range []string{"id", "label", "kind", "public_key", "public_key_base64", "is_default"} {
		if _, ok := elemResource.Schema[field]; !ok {
			t.Fatalf("expected field %q in keys elem schema", field)
		}
	}
}
