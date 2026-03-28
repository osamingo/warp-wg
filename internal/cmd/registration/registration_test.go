package registration_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/warp-wg/internal/cmd/registration"
	"github.com/osamingo/warp-wg/internal/warp"
)

func TestPromptTOS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "success: accepts y", input: "y\n"},
		{name: "success: accepts Y", input: "Y\n"},
		{name: "success: accepts yes", input: "yes\n"},
		{name: "success: accepts YES", input: "YES\n"},
		{name: "success: trims whitespace", input: "  y  \n"},
		{name: "error: rejects n", input: "n\n", wantErr: true},
		{name: "error: rejects empty", input: "\n", wantErr: true},
		{name: "error: rejects arbitrary text", input: "maybe\n", wantErr: true},
		{name: "error: empty reader (EOF)", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}

			err := registration.PromptTOS(in, out)

			if (err != nil) != tt.wantErr {
				t.Fatalf("PromptTOS() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !strings.Contains(out.String(), "cloudflare.com") {
				t.Error("prompt should contain ToS URL")
			}
		})
	}
}

func TestConfirmDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "success: accepts y", input: "y\n"},
		{name: "success: accepts yes", input: "yes\n"},
		{name: "error: rejects n", input: "n\n", wantErr: true},
		{name: "error: rejects empty", input: "\n", wantErr: true},
		{name: "error: empty reader (EOF)", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}

			err := registration.ConfirmDelete(in, out)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ConfirmDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSystemLocale(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want string
	}{
		{
			name: "success: returns LC_ALL locale",
			envs: map[string]string{"LC_ALL": "ja_JP.UTF-8"},
			want: "ja_JP",
		},
		{
			name: "success: LC_ALL takes precedence over LANG",
			envs: map[string]string{"LC_ALL": "fr_FR.UTF-8", "LANG": "en_US.UTF-8"},
			want: "fr_FR",
		},
		{
			name: "success: falls back to LC_MESSAGES",
			envs: map[string]string{"LC_MESSAGES": "de_DE.UTF-8"},
			want: "de_DE",
		},
		{
			name: "success: falls back to LANG",
			envs: map[string]string{"LANG": "ko_KR.EUC-KR"},
			want: "ko_KR",
		},
		{
			name: "success: returns value without encoding suffix",
			envs: map[string]string{"LANG": "C"},
			want: "C",
		},
		{
			name: "success: defaults to en_US",
			envs: map[string]string{},
			want: "en_US",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LC_ALL", "")
			t.Setenv("LC_MESSAGES", "")
			t.Setenv("LANG", "")

			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			if diff := cmp.Diff(tt.want, registration.SystemLocale()); diff != "" {
				t.Errorf("SystemLocale() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPrintDevice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		device *warp.DeviceResponse
		want   string
	}{
		{
			name: "success: prints all fields including peer",
			device: &warp.DeviceResponse{
				ID: "test-device-id",
				Account: warp.Account{
					AccountType: "free",
					PremiumData: 0,
					Quota:       0,
					Created:     "2026-03-29T00:00:00Z",
					Updated:     "2026-03-29T00:00:00Z",
				},
				Config: warp.DeviceConfig{
					Interface: struct {
						Addresses struct {
							V4 string `json:"v4"`
							V6 string `json:"v6"`
						} `json:"addresses"`
					}{
						Addresses: struct {
							V4 string `json:"v4"`
							V6 string `json:"v6"`
						}{V4: "172.16.0.2", V6: "fd01::1"},
					},
					Peers: []warp.Peer{{
						PublicKey: "server-pub-key",
						Endpoint: struct {
							Host string `json:"host"`
							V4   string `json:"v4"`
							V6   string `json:"v6"`
						}{Host: "engage.cloudflareclient.com"},
					}},
				},
			},
			want: "Device ID:    test-device-id\n" +
				"Account Type: free\n" +
				"Premium Data: 0 B\n" +
				"Quota:        0 B\n" +
				"Created:      2026-03-29T00:00:00Z\n" +
				"Updated:      2026-03-29T00:00:00Z\n" +
				"IPv4:         172.16.0.2\n" +
				"IPv6:         fd01::1\n" +
				"Endpoint:     engage.cloudflareclient.com\n" +
				"Public Key:   server-pub-key\n",
		},
		{
			name: "success: omits peer fields when no peers",
			device: &warp.DeviceResponse{
				ID: "no-peers",
				Account: warp.Account{
					AccountType: "limited",
					PremiumData: 1073741824,
					Quota:       10737418240,
					Created:     "2026-01-01T00:00:00Z",
					Updated:     "2026-03-01T00:00:00Z",
				},
			},
			want: "Device ID:    no-peers\n" +
				"Account Type: limited\n" +
				"Premium Data: 1.0 GiB\n" +
				"Quota:        10 GiB\n" +
				"Created:      2026-01-01T00:00:00Z\n" +
				"Updated:      2026-03-01T00:00:00Z\n" +
				"IPv4:         \n" +
				"IPv6:         \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			if err := registration.PrintDevice(&buf, tt.device); err != nil {
				t.Fatalf("PrintDevice() error = %v", err)
			}

			if diff := cmp.Diff(tt.want, buf.String()); diff != "" {
				t.Errorf("PrintDevice() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
