package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtraHeadersAreSentButCannotReplaceTheToken(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	c, err := NewClient(&Config{
		Host:  srv.URL + "/@warpgate/admin/api",
		Token: "real-token",
		Headers: map[string]string{
			"CF-Access-Client-Id": "client-id",
			"X-Warpgate-Token":    "spoofed",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := c.GetRoles(context.Background(), ""); err != nil {
		t.Fatal(err)
	}

	if v := got.Get("CF-Access-Client-Id"); v != "client-id" {
		t.Fatalf("expected the extra header to be sent, got %q", v)
	}
	if v := got.Get("X-Warpgate-Token"); v != "real-token" {
		t.Fatalf("expected the configured token to win, got %q", v)
	}
}
