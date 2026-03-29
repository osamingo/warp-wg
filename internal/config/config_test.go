package config_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/warp-wg/internal/config"
)

func TestSaveAndLoad(t *testing.T) {
	tests := []struct {
		name string
		acct *config.Registration
	}{
		{
			name: "success: saves and loads all fields",
			acct: &config.Registration{
				RegistrationID: "test-device-id",
				APIToken:       "test-access-token",
				PrivateKey:     "YAnezg1qdTdRLGL7F+FPBnEuIc/6vmNPiPxP0GG2GA0=",
			},
		},
		{
			name: "success: saves and loads with empty fields",
			acct: &config.Registration{
				RegistrationID: "device-only",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", dir)

			if err := config.Save(context.Background(), tt.acct); err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			got, err := config.Load(context.Background())
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if diff := cmp.Diff(tt.acct, got); diff != "" {
				t.Errorf("Load() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string)
		want    *config.Registration
		wantErr bool
	}{
		{
			name:  "success: returns empty account when file does not exist",
			setup: func(t *testing.T, dir string) { t.Helper() },
			want:  &config.Registration{},
		},
		{
			name: "error: returns error for invalid JSON",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				appDir := filepath.Join(dir, "warp-wg")
				if err := os.MkdirAll(appDir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(appDir, "reg.json"), []byte("{invalid json}"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", dir)
			tt.setup(t, dir)

			got, err := config.Load(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("Load() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	tests := []struct {
		name string
		file *config.Registration
		envs map[string]string
		want *config.Registration
	}{
		{
			name: "success: environment variables override file values",
			file: &config.Registration{
				RegistrationID: "file-device",
				APIToken:       "file-token",
				PrivateKey:     "file-key",
			},
			envs: map[string]string{
				"WARP_WG_REGISTRATION_ID": "env-device",
				"WARP_WG_API_TOKEN":       "env-token",
				"WARP_WG_PRIVATE_KEY":     "env-key",
			},
			want: &config.Registration{
				RegistrationID: "env-device",
				APIToken:       "env-token",
				PrivateKey:     "env-key",
			},
		},
		{
			name: "success: partial override keeps file values",
			file: &config.Registration{
				RegistrationID: "file-device",
				APIToken:       "file-token",
				PrivateKey:     "file-key",
			},
			envs: map[string]string{
				"WARP_WG_API_TOKEN": "env-token",
			},
			want: &config.Registration{
				RegistrationID: "file-device",
				APIToken:       "env-token",
				PrivateKey:     "file-key",
			},
		},
		{
			name: "success: environment variables work without file",
			file: nil,
			envs: map[string]string{
				"WARP_WG_REGISTRATION_ID": "env-device",
			},
			want: &config.Registration{
				RegistrationID: "env-device",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", dir)

			if tt.file != nil {
				if err := config.Save(context.Background(), tt.file); err != nil {
					t.Fatalf("Save() error = %v", err)
				}
			}

			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			got, err := config.Load(context.Background())
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Load() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep")
	t.Setenv("XDG_CONFIG_HOME", dir)

	acct := &config.Registration{RegistrationID: "test"}
	if err := config.Save(context.Background(), acct); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := filepath.Join(dir, "warp-wg", "reg.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file should exist at %s: %v", path, err)
	}
}

func TestSave_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	acct := &config.Registration{
		RegistrationID: "test",
		APIToken:       "secret-token",
		PrivateKey:     "secret-key",
	}
	if err := config.Save(context.Background(), acct); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := filepath.Join(dir, "warp-wg", "reg.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	wantPerm := os.FileMode(0o600)
	if diff := cmp.Diff(wantPerm, info.Mode().Perm()); diff != "" {
		t.Errorf("file permission mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadRegistered(t *testing.T) {
	tests := []struct {
		name    string
		acct    *config.Registration
		wantErr bool
	}{
		{
			name: "success: all fields present",
			acct: &config.Registration{
				RegistrationID: "device-id",
				APIToken:       "token",
				PrivateKey:     "key",
			},
		},
		{
			name:    "error: empty config",
			acct:    &config.Registration{},
			wantErr: true,
		},
		{
			name: "error: missing access_token",
			acct: &config.Registration{
				RegistrationID: "device-id",
				PrivateKey:     "key",
			},
			wantErr: true,
		},
		{
			name: "error: missing private_key",
			acct: &config.Registration{
				RegistrationID: "device-id",
				APIToken:       "token",
			},
			wantErr: true,
		},
		{
			name: "error: missing device_id",
			acct: &config.Registration{
				APIToken:   "token",
				PrivateKey: "key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", dir)
			ctx := context.Background()

			if err := config.Save(ctx, tt.acct); err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			_, err := config.LoadRegistered(ctx)

			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadRegistered() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithPath(t *testing.T) {
	t.Parallel()

	customPath := filepath.Join(t.TempDir(), "custom.toml")
	ctx := config.WithPath(context.Background(), customPath)

	acct := &config.Registration{
		RegistrationID: "test",
		APIToken:       "token",
		PrivateKey:     "key",
	}
	if err := config.Save(ctx, acct); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := config.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if diff := cmp.Diff(acct, got); diff != "" {
		t.Errorf("Load() mismatch (-want +got):\n%s", diff)
	}

	gotPath, err := config.FilePath(ctx)
	if err != nil {
		t.Fatalf("FilePath() error = %v", err)
	}
	if diff := cmp.Diff(customPath, gotPath); diff != "" {
		t.Errorf("FilePath() mismatch (-want +got):\n%s", diff)
	}
}

func TestAccount_LogValue(t *testing.T) {
	t.Parallel()

	acct := &config.Registration{
		RegistrationID: "device-123",
		APIToken:       "super-secret-token",
		PrivateKey:     "super-secret-key",
	}

	logOutput := acct.LogValue().String()

	if strings.Contains(logOutput, "super-secret-token") {
		t.Error("LogValue() should not contain access token")
	}
	if strings.Contains(logOutput, "super-secret-key") {
		t.Error("LogValue() should not contain private key")
	}
	if !strings.Contains(logOutput, "device-123") {
		t.Error("LogValue() should contain device ID")
	}
	if !strings.Contains(logOutput, "[REDACTED]") {
		t.Error("LogValue() should contain [REDACTED]")
	}

	// Verify it implements slog.LogValuer interface.
	var _ slog.LogValuer = acct
}
