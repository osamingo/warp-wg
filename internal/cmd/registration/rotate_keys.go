package registration

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
	"github.com/osamingo/warp-wg/internal/wireguard"
)

// NewRotateKeysCmd creates the "registration rotate-keys" command.
func NewRotateKeysCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("rotate-keys").SetParent(parentFlags)
	quiet := flags.Bool('q', "quiet", "Skip confirmation prompt")

	return &ff.Command{
		Name:      "rotate-keys",
		Usage:     "warp-wg registration rotate-keys [-q]",
		ShortHelp: "Generate a new key pair and update the registration",
		Flags:     flags,
		Exec: func(ctx context.Context, _ []string) error {
			if !*quiet {
				if err := confirmRotateKeys(os.Stdin, os.Stderr); err != nil {
					return err
				}
			}
			return execRotateKeys(ctx)
		},
	}
}

func confirmRotateKeys(in io.Reader, out io.Writer) error {
	if _, err := fmt.Fprint(out, "This will generate a new key pair and invalidate the current one. Continue? (y/N): "); err != nil {
		return fmt.Errorf("writing prompt: %w", err)
	}

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		return errors.New("no input received")
	}

	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if answer != "y" && answer != "yes" {
		return errors.New("aborted")
	}

	return nil
}

func execRotateKeys(ctx context.Context) error {
	reg, err := config.LoadRegistered(ctx)
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
	if _, err := client.UpdateDeviceKey(ctx, reg.RegistrationID, reg.APIToken, &warp.UpdateDeviceRequest{
		Key: pubKey.String(),
	}); err != nil {
		return fmt.Errorf("updating device key: %w", err)
	}

	reg.PrivateKey = privKey.String()
	if err := config.Save(ctx, reg); err != nil {
		slog.Error("failed to save config, manually save the following credentials",
			slog.String("private_key", privKey.String()),
		)
		return fmt.Errorf("remote key was updated but local config save failed, re-register may be required: %w", err)
	}

	slog.Info("keys rotated successfully")

	return nil
}
