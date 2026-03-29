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
		acct *config.Account
	}{
		{
			name: "success: saves and loads all fields",
			acct: &config.Account{
				DeviceID:    "test-device-id",
				AccessToken: "test-access-token",
				PrivateKey:  "YAnezg1qdTdRLGL7F+FPBnEuIc/6vmNPiPxP0GG2GA0=",
			},
		},
		{
			name: "success: saves and loads with empty fields",
			acct: &config.Account{
				DeviceID: "device-only",
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
		want    *config.Account
		wantErr bool
	}{
		{
			name:  "success: returns empty account when file does not exist",
			setup: func(t *testing.T, dir string) { t.Helper() },
			want:  &config.Account{},
		},
		{
			name: "error: returns error for invalid TOML",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				appDir := filepath.Join(dir, "warp-wg")
				if err := os.MkdirAll(appDir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(appDir, "account.toml"), []byte("invalid[[[toml"), 0o600); err != nil {
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
		file *config.Account
		envs map[string]string
		want *config.Account
	}{
		{
			name: "success: environment variables override file values",
			file: &config.Account{
				DeviceID:    "file-device",
				AccessToken: "file-token",
				PrivateKey:  "file-key",
			},
			envs: map[string]string{
				"WARP_WG_DEVICE_ID":    "env-device",
				"WARP_WG_ACCESS_TOKEN": "env-token",
				"WARP_WG_PRIVATE_KEY":  "env-key",
			},
			want: &config.Account{
				DeviceID:    "env-device",
				AccessToken: "env-token",
				PrivateKey:  "env-key",
			},
		},
		{
			name: "success: partial override keeps file values",
			file: &config.Account{
				DeviceID:    "file-device",
				AccessToken: "file-token",
				PrivateKey:  "file-key",
			},
			envs: map[string]string{
				"WARP_WG_ACCESS_TOKEN": "env-token",
			},
			want: &config.Account{
				DeviceID:    "file-device",
				AccessToken: "env-token",
				PrivateKey:  "file-key",
			},
		},
		{
			name: "success: environment variables work without file",
			file: nil,
			envs: map[string]string{
				"WARP_WG_DEVICE_ID": "env-device",
			},
			want: &config.Account{
				DeviceID: "env-device",
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

	acct := &config.Account{DeviceID: "test"}
	if err := config.Save(context.Background(), acct); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := filepath.Join(dir, "warp-wg", "account.toml")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file should exist at %s: %v", path, err)
	}
}

func TestSave_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	acct := &config.Account{
		DeviceID:    "test",
		AccessToken: "secret-token",
		PrivateKey:  "secret-key",
	}
	if err := config.Save(context.Background(), acct); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := filepath.Join(dir, "warp-wg", "account.toml")
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
		acct    *config.Account
		wantErr bool
	}{
		{
			name: "success: all fields present",
			acct: &config.Account{
				DeviceID:    "device-id",
				AccessToken: "token",
				PrivateKey:  "key",
			},
		},
		{
			name:    "error: empty config",
			acct:    &config.Account{},
			wantErr: true,
		},
		{
			name: "error: missing access_token",
			acct: &config.Account{
				DeviceID:   "device-id",
				PrivateKey: "key",
			},
			wantErr: true,
		},
		{
			name: "error: missing private_key",
			acct: &config.Account{
				DeviceID:    "device-id",
				AccessToken: "token",
			},
			wantErr: true,
		},
		{
			name: "error: missing device_id",
			acct: &config.Account{
				AccessToken: "token",
				PrivateKey:  "key",
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

	acct := &config.Account{
		DeviceID:    "test",
		AccessToken: "token",
		PrivateKey:  "key",
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

	acct := &config.Account{
		DeviceID:    "device-123",
		AccessToken: "super-secret-token",
		PrivateKey:  "super-secret-key",
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
