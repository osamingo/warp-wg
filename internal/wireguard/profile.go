package wireguard

import (
	"encoding/base64"
	"fmt"
	"io"
	"text/template"
)

var profileTmpl = template.Must(template.New("profile").Parse(`[Interface]
PrivateKey = {{ .PrivateKey }}
Address = {{ .Address }}
DNS = {{ .DNS }}
MTU = {{ .MTU }}

[Peer]
PublicKey = {{ .PeerPublicKey }}
AllowedIPs = {{ .AllowedIPs }}
Endpoint = {{ .PeerEndpoint }}
PersistentKeepalive = 25
{{- if .Reserved }}
{{ .ReservedPrefix }}Reserved = {{ .Reserved }}
{{- if not .ReservedActive }}
# Cloudflare client_id for load balancing (used by Xray-core, sing-box, etc.)
{{- end }}
{{- end }}
`))

// ProfileData holds the values for the WireGuard profile template.
type ProfileData struct {
	PrivateKey     string
	Address        string
	DNS            string
	MTU            int
	AllowedIPs     string
	PeerPublicKey  string
	PeerEndpoint   string
	Reserved       string
	ReservedPrefix string
	ReservedActive bool
}

// ProfileOptions controls the profile generation behavior.
type ProfileOptions struct {
	NoIPv6   bool
	MTU      int
	Reserved bool
}

// NewProfileData builds ProfileData from raw API values and options.
func NewProfileData(privKey, addrV4, addrV6, peerPubKey, peerEndpoint, reserved string, opts ProfileOptions) *ProfileData {
	address := addrV4 + "/32"
	dns := "1.1.1.1, 1.0.0.1"
	allowedIPs := "0.0.0.0/0"

	if !opts.NoIPv6 {
		address += ", " + addrV6 + "/128"
		dns += ", 2606:4700:4700::1111, 2606:4700:4700::1001"
		allowedIPs += ", ::/0"
	}

	reservedPrefix := "# "
	if opts.Reserved {
		reservedPrefix = ""
	}

	return &ProfileData{
		PrivateKey:     privKey,
		Address:        address,
		DNS:            dns,
		MTU:            opts.MTU,
		AllowedIPs:     allowedIPs,
		PeerPublicKey:  peerPubKey,
		PeerEndpoint:   peerEndpoint,
		Reserved:       reserved,
		ReservedPrefix: reservedPrefix,
		ReservedActive: opts.Reserved,
	}
}

// ClientIDToReserved converts a base64-encoded Cloudflare client_id to
// a comma-separated decimal byte string (e.g. "171, 85, 205").
func ClientIDToReserved(clientID string) (string, error) {
	if clientID == "" {
		return "", nil
	}

	b, err := base64.StdEncoding.DecodeString(clientID)
	if err != nil {
		return "", fmt.Errorf("decoding client_id: %w", err)
	}

	if len(b) < 3 {
		return "", fmt.Errorf("client_id too short: got %d bytes, want at least 3", len(b))
	}

	return fmt.Sprintf("%d, %d, %d", b[0], b[1], b[2]), nil
}

// WriteProfile writes a WireGuard configuration to the given writer.
func WriteProfile(out io.Writer, data *ProfileData) error {
	if err := profileTmpl.Execute(out, data); err != nil {
		return fmt.Errorf("writing profile: %w", err)
	}
	return nil
}
