package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTargetRoleFileTransferPermission(t *testing.T) {
	tests := []struct {
		name           string
		targetID       string
		roleID         string
		responseStatus int
		responseBody   *FileTransferPermission
		wantErr        bool
		wantNil        bool
	}{
		{
			name:           "successful get with all fields",
			targetID:       "target-123",
			roleID:         "role-456",
			responseStatus: http.StatusOK,
			responseBody: &FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: false,
				AllowedPaths:      []string{"/home", "/var/log"},
				BlockedExtensions: []string{".exe", ".sh"},
				MaxFileSize:       int64Ptr(1024),
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:           "successful get with defaults only",
			targetID:       "target-123",
			roleID:         "role-456",
			responseStatus: http.StatusOK,
			responseBody: &FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:           "not found returns nil",
			targetID:       "target-123",
			roleID:         "role-456",
			responseStatus: http.StatusNotFound,
			responseBody:   nil,
			wantErr:        false,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/targets/" + tt.targetID + "/roles/" + tt.roleID + "/file-transfer"
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: got %s, want %s", r.URL.Path, expectedPath)
				}
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method: got %s, want GET", r.Method)
				}

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				Host:  server.URL + "/api",
				Token: "test-token",
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			result, err := client.GetTargetRoleFileTransferPermission(context.Background(), tt.targetID, tt.roleID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTargetRoleFileTransferPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantNil && result != nil {
				t.Errorf("GetTargetRoleFileTransferPermission() = %v, want nil", result)
				return
			}

			if !tt.wantNil && result == nil {
				t.Error("GetTargetRoleFileTransferPermission() = nil, want non-nil")
				return
			}

			if result != nil && tt.responseBody != nil {
				if result.AllowFileUpload != tt.responseBody.AllowFileUpload {
					t.Errorf("AllowFileUpload = %v, want %v", result.AllowFileUpload, tt.responseBody.AllowFileUpload)
				}
				if result.AllowFileDownload != tt.responseBody.AllowFileDownload {
					t.Errorf("AllowFileDownload = %v, want %v", result.AllowFileDownload, tt.responseBody.AllowFileDownload)
				}
			}
		})
	}
}

func TestUpdateTargetRoleFileTransferPermission(t *testing.T) {
	tests := []struct {
		name           string
		targetID       string
		roleID         string
		request        *FileTransferPermission
		responseStatus int
		responseBody   *FileTransferPermission
		wantErr        bool
	}{
		{
			name:     "successful update with restrictions",
			targetID: "target-123",
			roleID:   "role-456",
			request: &FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: false,
				AllowedPaths:      []string{"/home/deploy"},
				BlockedExtensions: []string{".exe"},
				MaxFileSize:       int64Ptr(10485760),
			},
			responseStatus: http.StatusOK,
			responseBody: &FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: false,
				AllowedPaths:      []string{"/home/deploy"},
				BlockedExtensions: []string{".exe"},
				MaxFileSize:       int64Ptr(10485760),
			},
			wantErr: false,
		},
		{
			name:     "successful update to defaults",
			targetID: "target-123",
			roleID:   "role-456",
			request: &FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
			},
			responseStatus: http.StatusOK,
			responseBody: &FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
			},
			wantErr: false,
		},
		{
			name:     "not found error",
			targetID: "target-123",
			roleID:   "role-456",
			request: &FileTransferPermission{
				AllowFileUpload:   true,
				AllowFileDownload: true,
			},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/targets/" + tt.targetID + "/roles/" + tt.roleID + "/file-transfer"
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: got %s, want %s", r.URL.Path, expectedPath)
				}
				if r.Method != http.MethodPut {
					t.Errorf("unexpected method: got %s, want PUT", r.Method)
				}

				// Verify request body
				var reqBody FileTransferPermission
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}

				if reqBody.AllowFileUpload != tt.request.AllowFileUpload {
					t.Errorf("request AllowFileUpload = %v, want %v", reqBody.AllowFileUpload, tt.request.AllowFileUpload)
				}
				if reqBody.AllowFileDownload != tt.request.AllowFileDownload {
					t.Errorf("request AllowFileDownload = %v, want %v", reqBody.AllowFileDownload, tt.request.AllowFileDownload)
				}

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				Host:  server.URL + "/api",
				Token: "test-token",
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			result, err := client.UpdateTargetRoleFileTransferPermission(context.Background(), tt.targetID, tt.roleID, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTargetRoleFileTransferPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("UpdateTargetRoleFileTransferPermission() = nil, want non-nil")
				return
			}

			if result != nil && tt.responseBody != nil {
				if result.AllowFileUpload != tt.responseBody.AllowFileUpload {
					t.Errorf("AllowFileUpload = %v, want %v", result.AllowFileUpload, tt.responseBody.AllowFileUpload)
				}
				if result.AllowFileDownload != tt.responseBody.AllowFileDownload {
					t.Errorf("AllowFileDownload = %v, want %v", result.AllowFileDownload, tt.responseBody.AllowFileDownload)
				}
			}
		})
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}
