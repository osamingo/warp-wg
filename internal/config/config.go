package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	appName  = "warp-wg"
	fileName = "reg.json"

	envPrefix         = "WARP_WG_"
	envRegistrationID = envPrefix + "REGISTRATION_ID"
	envAPIToken       = envPrefix + "API_TOKEN"
	envPrivateKey     = envPrefix + "PRIVATE_KEY"
)

// ErrNotRegistered is returned when no device registration is found.
var ErrNotRegistered = errors.New("not registered, run 'warp-wg registration new' first")

// Registration holds the WARP device registration credentials.
type Registration struct {
	RegistrationID string `json:"registration_id"`
	APIToken       string `json:"api_token"`
	PrivateKey     string `json:"private_key"`
}

// LogValue implements slog.LogValuer to prevent secrets from being logged.
func (r *Registration) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("registration_id", r.RegistrationID),
		slog.String("api_token", "[REDACTED]"),
		slog.String("private_key", "[REDACTED]"),
	)
}

type pathKey struct{}

// WithPath returns a context that carries a custom config file path.
func WithPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, pathKey{}, path)
}

// Load reads the registration from the JSON file,
// then applies any environment variable overrides.
func Load(ctx context.Context) (*Registration, error) {
	path, err := FilePath(ctx)
	if err != nil {
		return nil, err
	}

	var reg Registration

	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err == nil {
		if err := json.Unmarshal(data, &reg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	applyEnvOverrides(&reg)

	return &reg, nil
}

// LoadRegistered loads the registration and returns ErrNotRegistered if no
// device registration exists.
func LoadRegistered(ctx context.Context) (*Registration, error) {
	reg, err := Load(ctx)
	if err != nil {
		return nil, err
	}
	if reg.RegistrationID == "" || reg.APIToken == "" || reg.PrivateKey == "" {
		return nil, ErrNotRegistered
	}
	return reg, nil
}

// Save writes the registration to the JSON file.
func Save(ctx context.Context, reg *Registration) error {
	path, err := FilePath(ctx)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(reg, "", "  ") //nolint:gosec // credentials are intentionally saved to a file with 0600 permissions
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// FilePath returns the path to the registration file.
// If a custom path was set via WithPath, it is returned instead.
func FilePath(ctx context.Context) (string, error) {
	if path, ok := ctx.Value(pathKey{}).(string); ok && path != "" {
		return path, nil
	}

	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName, fileName), nil
}

func configDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}

	return filepath.Join(home, ".config"), nil
}

func applyEnvOverrides(reg *Registration) {
	if v := os.Getenv(envRegistrationID); v != "" {
		reg.RegistrationID = v
	}
	if v := os.Getenv(envAPIToken); v != "" {
		reg.APIToken = v
	}
	if v := os.Getenv(envPrivateKey); v != "" {
		reg.PrivateKey = v
	}
}
