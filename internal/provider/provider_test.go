package provider

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// Acceptance tests run against a live Warpgate. They are skipped unless TF_ACC
// is set, and need WARPGATE_HOST and WARPGATE_TOKEN pointing at an instance
// whose state they are free to overwrite. scripts/test-warpgate.sh starts a
// throwaway one.
var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"warpgate": func() (*schema.Provider, error) {
		return New("acctest")(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New("test")().InternalValidate(); err != nil {
		t.Fatalf("provider schema is invalid: %v", err)
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	for _, key := range []string{"WARPGATE_HOST", "WARPGATE_TOKEN"} {
		if os.Getenv(key) == "" {
			t.Fatalf("%s must be set for acceptance tests", key)
		}
	}
}

// testAccClient talks to the same instance the provider under test does, so a
// check can assert what actually reached Warpgate rather than only what the
// provider wrote to state.
func testAccClient(t *testing.T) *client.Client {
	t.Helper()

	host := os.Getenv("WARPGATE_HOST")

	apiPath := "/@warpgate/admin/api"
	if !strings.Contains(host, apiPath) {
		host = strings.TrimSuffix(host, "/") + apiPath
	}

	c, err := client.NewClient(&client.Config{
		Host:               host,
		Token:              os.Getenv("WARPGATE_TOKEN"),
		InsecureSkipVerify: os.Getenv("WARPGATE_INSECURE_SKIP_VERIFY") != "",
	})
	if err != nil {
		t.Fatalf("failed to build the check client: %v", err)
	}

	return c
}

func testAccParameters(t *testing.T) *client.ParameterValues {
	t.Helper()

	params, err := testAccClient(t).GetParameters(context.Background())
	if err != nil {
		t.Fatalf("failed to read parameters: %v", err)
	}

	return params
}
