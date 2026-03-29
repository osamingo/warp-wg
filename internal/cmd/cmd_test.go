package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/warp-wg/internal/cmd"
	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
)

func TestPrintCompletion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		shell   string
		wantErr bool
		contain string
	}{
		{
			name:    "success: bash completion",
			shell:   "bash",
			contain: "complete -F _warp_wg warp-wg",
		},
		{
			name:    "success: zsh completion",
			shell:   "zsh",
			contain: "#compdef warp-wg",
		},
		{
			name:    "success: fish completion",
			shell:   "fish",
			contain: "complete -c warp-wg",
		},
		{
			name:    "error: unsupported shell",
			shell:   "powershell",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := cmd.PrintCompletion(&buf, tt.shell)

			if (err != nil) != tt.wantErr {
				t.Fatalf("PrintCompletion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !strings.Contains(buf.String(), tt.contain) {
				t.Errorf("PrintCompletion(%q) should contain %q", tt.shell, tt.contain)
			}
		})
	}
}

func TestPeerEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		v4   string
		port string
		want string
	}{
		{
			name: "success: strips port 0 and adds specified port",
			v4:   "162.159.192.1:0",
			port: "2408",
			want: "162.159.192.1:2408",
		},
		{
			name: "success: replaces existing port",
			v4:   "162.159.192.1:51820",
			port: "2408",
			want: "162.159.192.1:2408",
		},
		{
			name: "success: adds port when none present",
			v4:   "162.159.192.1",
			port: "2408",
			want: "162.159.192.1:2408",
		},
		{
			name: "success: handles IPv6 with port",
			v4:   "[2606:4700::1]:0",
			port: "2408",
			want: "[2606:4700::1]:2408",
		},
		{
			name: "success: custom port",
			v4:   "162.159.192.1:0",
			port: "500",
			want: "162.159.192.1:500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(tt.want, cmd.PeerEndpoint(tt.v4, tt.port)); diff != "" {
				t.Errorf("PeerEndpoint() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExecProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		flags   cmd.ProfileFlags
		handler http.HandlerFunc
		wantErr bool
		contain string
	}{
		{
			name:  "success: generates profile with defaults",
			flags: cmd.ProfileFlags{MTU: 1420, Port: "2408"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"id": "device-123",
					"account": map[string]any{
						"id":           "acct-123",
						"account_type": "free",
					},
					"config": map[string]any{
						"client_id": "AQID",
						"interface": map[string]any{
							"addresses": map[string]string{
								"v4": "172.16.0.2",
								"v6": "fd01::1",
							},
						},
						"peers": []map[string]any{{
							"public_key": "server-pub-key",
							"endpoint": map[string]string{
								"host": "engage.cloudflareclient.com:2408",
								"v4":   "162.159.192.1:0",
							},
						}},
					},
				})
			},
			contain: "Endpoint = engage.cloudflareclient.com:2408",
		},
		{
			name:  "success: no-ipv6 excludes IPv6",
			flags: cmd.ProfileFlags{NoIPv6: true, MTU: 1420, Port: "2408"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"id":      "device-123",
					"account": map[string]any{"account_type": "free"},
					"config": map[string]any{
						"interface": map[string]any{
							"addresses": map[string]string{"v4": "172.16.0.2", "v6": "fd01::1"},
						},
						"peers": []map[string]any{{
							"public_key": "key",
							"endpoint":   map[string]string{"host": "host:2408", "v4": "1.2.3.4:0"},
						}},
					},
				})
			},
			contain: "Address = 172.16.0.2/32\n",
		},
		{
			name:  "success: endpoint-ip uses IP",
			flags: cmd.ProfileFlags{UseIP: true, MTU: 1420, Port: "2408"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"id":      "device-123",
					"account": map[string]any{"account_type": "free"},
					"config": map[string]any{
						"interface": map[string]any{
							"addresses": map[string]string{"v4": "172.16.0.2", "v6": "fd01::1"},
						},
						"peers": []map[string]any{{
							"public_key": "key",
							"endpoint":   map[string]string{"host": "host:2408", "v4": "162.159.192.1:0"},
						}},
					},
				})
			},
			contain: "Endpoint = 162.159.192.1:2408",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			dir := t.TempDir()
			ctx := context.Background()
			ctx = config.WithPath(ctx, filepath.Join(dir, "account.toml"))
			ctx = warp.WithAPIBaseURL(ctx, srv.URL)

			acct := &config.Registration{
				RegistrationID: "device-123",
				APIToken:       "token",
				PrivateKey:     "YAnezg1qdTdRLGL7F+FPBnEuIc/6vmNPiPxP0GG2GA0=",
			}
			if err := config.Save(ctx, acct); err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			var buf bytes.Buffer
			err := cmd.ExecProfile(ctx, &buf, tt.flags)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ExecProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !strings.Contains(buf.String(), tt.contain) {
				t.Errorf("ExecProfile() should contain %q, got:\n%s", tt.contain, buf.String())
			}
		})
	}
}

func TestExecStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
		contain string
	}{
		{
			name: "success: returns trace output",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
					t.Errorf("method mismatch (-want +got):\n%s", diff)
				}
				w.Write([]byte("fl=test\nh=cloudflare.com\nip=1.2.3.4\n"))
			},
			contain: "ip=1.2.3.4",
		},
		{
			name: "error: server returns 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			var buf bytes.Buffer
			err := cmd.ExecStatus(context.Background(), &buf, srv.URL)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ExecStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !strings.Contains(buf.String(), tt.contain) {
				t.Errorf("ExecStatus() should contain %q, got %q", tt.contain, buf.String())
			}
		})
	}
}
