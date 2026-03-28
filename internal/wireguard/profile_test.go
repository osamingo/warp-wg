package wireguard_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/warp-wg/internal/wireguard"
)

func TestWriteProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    *wireguard.ProfileData
		want    string
		wantErr bool
	}{
		{
			name: "success: generates config with IPv6 and reserved",
			data: wireguard.NewProfileData(
				"YAnezg1qdTdRLGL7F+FPBnEuIc/6vmNPiPxP0GG2GA0=",
				"172.16.0.2",
				"2606:4700:110:8588:cf61:b4df:57f9:2198",
				"bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
				"engage.cloudflareclient.com:2408",
				"171, 85, 205",
				wireguard.ProfileOptions{MTU: 1420},
			),
			want: `[Interface]
PrivateKey = YAnezg1qdTdRLGL7F+FPBnEuIc/6vmNPiPxP0GG2GA0=
Address = 172.16.0.2/32, 2606:4700:110:8588:cf61:b4df:57f9:2198/128
DNS = 1.1.1.1, 1.0.0.1, 2606:4700:4700::1111, 2606:4700:4700::1001
MTU = 1420

[Peer]
PublicKey = bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = engage.cloudflareclient.com:2408
PersistentKeepalive = 25
# Reserved = 171, 85, 205
# Cloudflare client_id for load balancing (used by Xray-core, sing-box, etc.)
`,
		},
		{
			name: "success: generates config without IPv6 and custom MTU",
			data: wireguard.NewProfileData(
				"YAnezg1qdTdRLGL7F+FPBnEuIc/6vmNPiPxP0GG2GA0=",
				"172.16.0.2",
				"2606:4700:110:8588:cf61:b4df:57f9:2198",
				"bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
				"engage.cloudflareclient.com:2408",
				"",
				wireguard.ProfileOptions{NoIPv6: true, MTU: 1280},
			),
			want: `[Interface]
PrivateKey = YAnezg1qdTdRLGL7F+FPBnEuIc/6vmNPiPxP0GG2GA0=
Address = 172.16.0.2/32
DNS = 1.1.1.1, 1.0.0.1
MTU = 1280

[Peer]
PublicKey = bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=
AllowedIPs = 0.0.0.0/0
Endpoint = engage.cloudflareclient.com:2408
PersistentKeepalive = 25
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := wireguard.WriteProfile(&buf, tt.data)

			if (err != nil) != tt.wantErr {
				t.Fatalf("WriteProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, buf.String()); diff != "" {
					t.Errorf("WriteProfile() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestClientIDToReserved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		clientID string
		want     string
		wantErr  bool
	}{
		{
			name:     "success: converts 4-char base64 to 3 bytes",
			clientID: "q1XN",
			want:     "171, 85, 205",
		},
		{
			name:     "success: converts padded base64",
			clientID: "AQID",
			want:     "1, 2, 3",
		},
		{
			name:     "success: empty client_id returns empty string",
			clientID: "",
			want:     "",
		},
		{
			name:     "error: invalid base64",
			clientID: "!!!invalid!!!",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := wireguard.ClientIDToReserved(tt.clientID)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ClientIDToReserved() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("ClientIDToReserved() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
