package warp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/warp-wg/internal/warp"
)

func TestClient_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantID  string
		wantErr bool
	}{
		{
			name: "success: registers a new device",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if want := "/v0a5101/reg"; r.URL.Path != want {
					t.Errorf("path = %s, want %s", r.URL.Path, want)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
				}
				assertCommonHeaders(t, r)

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(warp.RegisterResponse{
					ID:    "device-123",
					Token: "token-abc",
				})
			},
			wantID: "device-123",
		},
		{
			name: "error: server returns 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal"}`))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			client := warp.NewClient(warp.WithBaseURL(srv.URL))
			got, err := client.Register(context.Background(), &warp.RegisterRequest{Key: "test"})

			if (err != nil) != tt.wantErr {
				t.Fatalf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("Register() ID = %s, want %s", got.ID, tt.wantID)
			}
		})
	}
}

func TestClient_Registration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "success: retrieves registration info",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if want := "/v0a5101/reg/device-123"; r.URL.Path != want {
					t.Errorf("path = %s, want %s", r.URL.Path, want)
				}
				assertAuth(t, r, "test-token")
				assertCommonHeaders(t, r)

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(warp.RegistrationResponse{ID: "device-123"})
			},
		},
		{
			name: "error: unauthorized",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized"}`))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			client := warp.NewClient(warp.WithBaseURL(srv.URL))
			_, err := client.Registration(context.Background(), "device-123", "test-token")

			if (err != nil) != tt.wantErr {
				t.Fatalf("Registration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeleteRegistration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "success: deletes registration",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("method = %s, want DELETE", r.Method)
				}
				assertAuth(t, r, "test-token")
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "error: not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":"not found"}`))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			client := warp.NewClient(warp.WithBaseURL(srv.URL))
			err := client.DeleteRegistration(context.Background(), "device-123", "test-token")

			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_UpdateRegistrationKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "success: updates registration key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("method = %s, want PATCH", r.Method)
				}
				assertAuth(t, r, "test-token")

				var req warp.UpdateRegistrationRequest
				json.NewDecoder(r.Body).Decode(&req)
				if req.Key != "new-public-key" {
					t.Errorf("key = %s, want new-public-key", req.Key)
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(warp.RegistrationResponse{ID: "device-123"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			client := warp.NewClient(warp.WithBaseURL(srv.URL))
			_, err := client.UpdateRegistrationKey(context.Background(), "device-123", "test-token", &warp.UpdateRegistrationRequest{Key: "new-public-key"})

			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateRegistrationKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_UpdateAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "success: updates license",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("method = %s, want PUT", r.Method)
				}
				assertAuth(t, r, "test-token")

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(warp.Account{AccountType: "unlimited"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			client := warp.NewClient(warp.WithBaseURL(srv.URL))
			got, err := client.UpdateAccount(context.Background(), "device-123", "test-token", &warp.UpdateAccountRequest{License: "key-123"})

			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.AccountType != "unlimited" {
				t.Errorf("AccountType = %s, want unlimited", got.AccountType)
			}
		})
	}
}

func TestClient_BoundDevices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    int
		wantErr bool
	}{
		{
			name: "success: returns device list",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				assertAuth(t, r, "test-token")

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]warp.BoundDevice{
					{ID: "dev-1", Name: "Phone", Active: true},
					{ID: "dev-2", Name: "PC", Active: false},
				})
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			client := warp.NewClient(warp.WithBaseURL(srv.URL))
			got, err := client.BoundDevices(context.Background(), "device-123", "test-token")

			if (err != nil) != tt.wantErr {
				t.Fatalf("BoundDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, len(got)); diff != "" {
					t.Errorf("BoundDevices() count mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()

	client := warp.NewClient(warp.WithBaseURL(srv.URL))
	_, err := client.Registration(context.Background(), "x", "t")

	var apiErr *warp.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error should be *warp.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}

func assertCommonHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Header.Get("User-Agent") != "okhttp/4.12.0" {
		t.Errorf("User-Agent = %s, want okhttp/4.12.0", r.Header.Get("User-Agent"))
	}
	if r.Header.Get("CF-Client-Version") != "a-6.38.6-5101" {
		t.Errorf("CF-Client-Version = %s, want a-6.38.6-5101", r.Header.Get("CF-Client-Version"))
	}
}

func assertAuth(t *testing.T, r *http.Request, token string) {
	t.Helper()
	want := "Bearer " + token
	if r.Header.Get("Authorization") != want {
		t.Errorf("Authorization = %s, want %s", r.Header.Get("Authorization"), want)
	}
}
