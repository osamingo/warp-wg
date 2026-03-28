package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const (
	appName  = "warp-wg"
	fileName = "account.toml"

	envPrefix     = "WARP_WG_"
	envDeviceID   = envPrefix + "DEVICE_ID"
	envToken      = envPrefix + "ACCESS_TOKEN"
	envPrivateKey = envPrefix + "PRIVATE_KEY"
)

// ErrNotRegistered is returned when no device registration is found.
var ErrNotRegistered = errors.New("not registered, run 'warp-wg registration new' first")

// Account holds the WARP device registration credentials.
type Account struct {
	DeviceID    string `toml:"device_id"`
	AccessToken string `toml:"access_token"`
	PrivateKey  string `toml:"private_key"`
}

type pathKey struct{}

// WithPath returns a context that carries a custom config file path.
func WithPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, pathKey{}, path)
}

// Load reads the account configuration from the TOML file,
// then applies any environment variable overrides.
func Load(ctx context.Context) (*Account, error) {
	path, err := FilePath(ctx)
	if err != nil {
		return nil, err
	}

	var acct Account

	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err == nil {
		if err := toml.Unmarshal(data, &acct); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	applyEnvOverrides(&acct)

	return &acct, nil
}

// LoadRegistered loads the account and returns ErrNotRegistered if no
// device registration exists.
func LoadRegistered(ctx context.Context) (*Account, error) {
	acct, err := Load(ctx)
	if err != nil {
		return nil, err
	}
	if acct.DeviceID == "" || acct.AccessToken == "" || acct.PrivateKey == "" {
		return nil, ErrNotRegistered
	}
	return acct, nil
}

// Save writes the account configuration to the TOML file.
func Save(ctx context.Context, acct *Account) error {
	path, err := FilePath(ctx)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := toml.Marshal(acct) //nolint:gosec // credentials are intentionally saved to a file with 0600 permissions
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// FilePath returns the path to the account configuration file.
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

func applyEnvOverrides(acct *Account) {
	if v := os.Getenv(envDeviceID); v != "" {
		acct.DeviceID = v
	}
	if v := os.Getenv(envToken); v != "" {
		acct.AccessToken = v
	}
	if v := os.Getenv(envPrivateKey); v != "" {
		acct.PrivateKey = v
	}
}
