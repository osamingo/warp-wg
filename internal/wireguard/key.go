package wireguard

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const keySize = 32

// Key represents a WireGuard key (Curve25519, 32 bytes).
type Key [keySize]byte

// GeneratePrivateKey generates a new random WireGuard private key.
func GeneratePrivateKey() (Key, error) {
	ecdhKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return Key{}, fmt.Errorf("generating private key: %w", err)
	}

	var key Key
	copy(key[:], ecdhKey.Bytes())

	return key, nil
}

// PublicKey derives the corresponding public key from a private key.
func (k *Key) PublicKey() (Key, error) {
	ecdhPriv, err := ecdh.X25519().NewPrivateKey(k[:])
	if err != nil {
		return Key{}, fmt.Errorf("deriving public key: %w", err)
	}

	var pub Key
	copy(pub[:], ecdhPriv.PublicKey().Bytes())

	return pub, nil
}

// String returns the base64-encoded representation of the key.
func (k *Key) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

// ParseKey decodes a base64-encoded WireGuard key.
func ParseKey(s string) (Key, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return Key{}, fmt.Errorf("decoding key: %w", err)
	}

	if len(b) != keySize {
		return Key{}, fmt.Errorf("invalid key length: got %d, want %d", len(b), keySize)
	}

	var key Key
	copy(key[:], b)

	return key, nil
}
