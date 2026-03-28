package wireguard_test

import (
	"encoding/base64"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/osamingo/warp-wg/internal/wireguard"
)

func TestGeneratePrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		check   func(t *testing.T, key wireguard.Key)
		wantErr bool
	}{
		{
			name: "success: generates a 32-byte key",
			check: func(t *testing.T, key wireguard.Key) {
				t.Helper()
				if key == (wireguard.Key{}) {
					t.Error("generated key should not be zero value")
				}
			},
		},
		{
			name: "success: generates unique keys",
			check: func(t *testing.T, key wireguard.Key) {
				t.Helper()
				key2, err := wireguard.GeneratePrivateKey()
				if err != nil {
					t.Fatalf("generating second key: %v", err)
				}
				if key == key2 {
					t.Error("two generated keys should not be equal")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := wireguard.GeneratePrivateKey()
			if (err != nil) != tt.wantErr {
				t.Fatalf("GeneratePrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestKey_PublicKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		privKey string
		wantPub string
		wantErr bool
	}{
		{
			name:    "success: derives correct public key from known private key",
			privKey: "YAnezg1qdTdRLGL7F+FPBnEuIc/6vmNPiPxP0GG2GA0=",
			wantPub: "wT+QsdD8F9/WMTERd6D99Soz00esEZ6T3ToDSKQ/9kI=",
		},
		{
			name:    "success: deterministic derivation",
			privKey: "OO2mBuIFmEKZk9/9TIuISN7ted1OtHWiXYbBCYzjy10=",
			wantPub: "bVIiSLZaF5xHSzlxGgTdnJjGSZgyN9kCZflC193GFSM=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			privKey, err := wireguard.ParseKey(tt.privKey)
			if err != nil {
				t.Fatalf("ParseKey() error = %v", err)
			}

			got, err := privKey.PublicKey()
			if (err != nil) != tt.wantErr {
				t.Fatalf("PublicKey() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.wantPub, got.String()); diff != "" {
				t.Errorf("PublicKey() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestKey_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  wireguard.Key
		want string
	}{
		{
			name: "success: encodes to base64 with 44 chars",
			key:  wireguard.Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			want: "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA=",
		},
		{
			name: "success: zero key encodes correctly",
			key:  wireguard.Key{},
			want: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(tt.want, tt.key.String()); diff != "" {
				t.Errorf("String() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    wireguard.Key
		wantErr bool
	}{
		{
			name:  "success: parses valid base64 key",
			input: "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA=",
			want:  wireguard.Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		},
		{
			name:    "error: invalid base64",
			input:   "not-valid-base64!!!",
			wantErr: true,
		},
		{
			name:    "error: wrong length",
			input:   base64.StdEncoding.EncodeToString([]byte("too short")),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := wireguard.ParseKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("ParseKey() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestKey_Roundtrip(t *testing.T) {
	t.Parallel()

	privKey, err := wireguard.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey() error = %v", err)
	}

	parsed, err := wireguard.ParseKey(privKey.String())
	if err != nil {
		t.Fatalf("ParseKey() error = %v", err)
	}

	if diff := cmp.Diff(privKey, parsed); diff != "" {
		t.Errorf("roundtrip mismatch (-want +got):\n%s", diff)
	}

	pub1, err := privKey.PublicKey()
	if err != nil {
		t.Fatalf("PublicKey() from original error = %v", err)
	}

	pub2, err := parsed.PublicKey()
	if err != nil {
		t.Fatalf("PublicKey() from parsed error = %v", err)
	}

	if diff := cmp.Diff(pub1, pub2); diff != "" {
		t.Errorf("public key mismatch after roundtrip (-want +got):\n%s", diff)
	}
}
