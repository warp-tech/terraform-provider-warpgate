package provider

import (
	"testing"

	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func TestBuildFileTransferPermission(t *testing.T) {
	tests := []struct {
		name     string
		opts     map[string]any
		expected *client.FileTransferPermission
	}{
		{
			name: "all fields set",
			opts: map[string]any{
				"allow_upload":       false,
				"allow_download":     false,
				"allowed_paths":      []any{"/home/deploy", "/var/log"},
				"blocked_extensions": []any{".exe", ".sh"},
				"max_file_size":      10485760,
			},
			expected: &client.FileTransferPermission{
				AllowFileUpload:   false,
				AllowFileDownload: false,
				AllowedPaths:      []string{"/home/deploy", "/var/log"},
				BlockedExtensions: []string{".exe", ".sh"},
				MaxFileSize:       int64Ptr(10485760),
			},
		},
		{
			name: "defaults only",
			opts: map[string]any{
				"allow_upload":   true,
				"allow_download": true,
			},
			expected: &client.FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
				AllowedPaths:      nil,
				BlockedExtensions: nil,
				MaxFileSize:       nil,
			},
		},
		{
			name: "upload only",
			opts: map[string]any{
				"allow_upload":   true,
				"allow_download": false,
			},
			expected: &client.FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: false,
				AllowedPaths:      nil,
				BlockedExtensions: nil,
				MaxFileSize:       nil,
			},
		},
		{
			name: "download only",
			opts: map[string]any{
				"allow_upload":   false,
				"allow_download": true,
			},
			expected: &client.FileTransferPermission{
				AllowFileUpload:   false,
				AllowFileDownload: true,
				AllowedPaths:      nil,
				BlockedExtensions: nil,
				MaxFileSize:       nil,
			},
		},
		{
			name: "with path restrictions",
			opts: map[string]any{
				"allow_upload":   true,
				"allow_download": true,
				"allowed_paths":  []any{"/home/user"},
			},
			expected: &client.FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
				AllowedPaths:      []string{"/home/user"},
				BlockedExtensions: nil,
				MaxFileSize:       nil,
			},
		},
		{
			name: "with extension restrictions",
			opts: map[string]any{
				"allow_upload":       true,
				"allow_download":     true,
				"blocked_extensions": []any{".sql", ".dump"},
			},
			expected: &client.FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
				AllowedPaths:      nil,
				BlockedExtensions: []string{".sql", ".dump"},
				MaxFileSize:       nil,
			},
		},
		{
			name: "with size limit",
			opts: map[string]any{
				"allow_upload":   true,
				"allow_download": true,
				"max_file_size":  52428800,
			},
			expected: &client.FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
				AllowedPaths:      nil,
				BlockedExtensions: nil,
				MaxFileSize:       int64Ptr(52428800),
			},
		},
		{
			name: "empty paths and extensions",
			opts: map[string]any{
				"allow_upload":       true,
				"allow_download":     true,
				"allowed_paths":      []any{},
				"blocked_extensions": []any{},
				"max_file_size":      0,
			},
			expected: &client.FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
				AllowedPaths:      nil,
				BlockedExtensions: nil,
				MaxFileSize:       nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFileTransferPermission(tt.opts)

			if result.AllowFileUpload != tt.expected.AllowFileUpload {
				t.Errorf("AllowFileUpload = %v, want %v", result.AllowFileUpload, tt.expected.AllowFileUpload)
			}

			if result.AllowFileDownload != tt.expected.AllowFileDownload {
				t.Errorf("AllowFileDownload = %v, want %v", result.AllowFileDownload, tt.expected.AllowFileDownload)
			}

			if !stringSliceEqual(result.AllowedPaths, tt.expected.AllowedPaths) {
				t.Errorf("AllowedPaths = %v, want %v", result.AllowedPaths, tt.expected.AllowedPaths)
			}

			if !stringSliceEqual(result.BlockedExtensions, tt.expected.BlockedExtensions) {
				t.Errorf("BlockedExtensions = %v, want %v", result.BlockedExtensions, tt.expected.BlockedExtensions)
			}

			if !int64PtrEqual(result.MaxFileSize, tt.expected.MaxFileSize) {
				t.Errorf("MaxFileSize = %v, want %v", ptrToString(result.MaxFileSize), ptrToString(tt.expected.MaxFileSize))
			}
		})
	}
}

func TestBuildFileTransferPermission_MissingFields(t *testing.T) {
	// Test with empty map - should use defaults
	opts := map[string]any{}
	result := buildFileTransferPermission(opts)

	if result.AllowFileUpload != true {
		t.Errorf("AllowFileUpload = %v, want true (default)", result.AllowFileUpload)
	}

	if result.AllowFileDownload != true {
		t.Errorf("AllowFileDownload = %v, want true (default)", result.AllowFileDownload)
	}

	if result.AllowedPaths != nil {
		t.Errorf("AllowedPaths = %v, want nil", result.AllowedPaths)
	}

	if result.BlockedExtensions != nil {
		t.Errorf("BlockedExtensions = %v, want nil", result.BlockedExtensions)
	}

	if result.MaxFileSize != nil {
		t.Errorf("MaxFileSize = %v, want nil", result.MaxFileSize)
	}
}

// Helper functions

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func int64Ptr(i int64) *int64 {
	return &i
}

func int64PtrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrToString(p *int64) string {
	if p == nil {
		return "nil"
	}
	return string(rune(*p))
}
