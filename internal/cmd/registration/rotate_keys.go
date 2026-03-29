package registration

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
	"github.com/osamingo/warp-wg/internal/wireguard"
)

// NewRotateKeysCmd creates the "registration rotate-keys" command.
func NewRotateKeysCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("rotate-keys").SetParent(parentFlags)

	return &ff.Command{
		Name:      "rotate-keys",
		Usage:     "warp-wg registration rotate-keys",
		ShortHelp: "Generate a new key pair and update the registration",
		Flags:     flags,
		Exec: func(ctx context.Context, _ []string) error {
			return execRotateKeys(ctx)
		},
	}
}

func execRotateKeys(ctx context.Context) error {
	acct, err := config.LoadRegistered(ctx)
	if err != nil {
		return err
	}

	privKey, err := wireguard.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("generating key pair: %w", err)
	}

	pubKey, err := privKey.PublicKey()
	if err != nil {
		return fmt.Errorf("deriving public key: %w", err)
	}

	slog.Info("rotating WireGuard keys")

	client := warp.NewClientFromContext(ctx)
	if _, err := client.UpdateDeviceKey(ctx, acct.DeviceID, acct.AccessToken, &warp.UpdateDeviceRequest{
		Key: pubKey.String(),
	}); err != nil {
		return fmt.Errorf("updating device key: %w", err)
	}

	acct.PrivateKey = privKey.String()
	if err := config.Save(ctx, acct); err != nil {
		return fmt.Errorf("remote key was updated but local config save failed, re-register may be required: %w", err)
	}

	slog.Info("keys rotated successfully")

	return nil
}
