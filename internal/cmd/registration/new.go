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
	"time"

	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
	"github.com/osamingo/warp-wg/internal/wireguard"
)

const tosURL = "https://www.cloudflare.com/application/terms/"

// NewNewCmd creates the "registration new" command.
func NewNewCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("new").SetParent(parentFlags)
	acceptTos := flags.Bool(0, "accept-tos", "Accept the Cloudflare Terms of Service without prompting")

	return &ff.Command{
		Name:      "new",
		Usage:     "warp-wg registration new [--accept-tos]",
		ShortHelp: "Register a new WARP device",
		Flags:     flags,
		Exec: func(ctx context.Context, _ []string) error {
			if !*acceptTos {
				if err := promptTOS(os.Stdin, os.Stderr); err != nil {
					return err
				}
			}
			return execNew(ctx)
		},
	}
}

func promptTOS(in io.Reader, out io.Writer) error {
	prompt := fmt.Sprintf(
		"This project is not affiliated with Cloudflare.\nTerms of Service: %s\nDo you agree? (y/N): ",
		tosURL,
	)
	if _, err := fmt.Fprint(out, prompt); err != nil {
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
		return errors.New("you must accept the Terms of Service to continue")
	}

	return nil
}

func execNew(ctx context.Context) error {
	acct, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if acct.DeviceID != "" {
		return fmt.Errorf("already registered (device_id: %s), delete the config file to re-register", acct.DeviceID)
	}

	privKey, err := wireguard.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("generating key pair: %w", err)
	}

	pubKey, err := privKey.PublicKey()
	if err != nil {
		return fmt.Errorf("deriving public key: %w", err)
	}

	slog.Info("registering device with Cloudflare WARP")

	client := warp.NewClientFromContext(ctx)
	resp, err := client.Register(ctx, &warp.RegisterRequest{
		Key:          pubKey.String(),
		InstallID:    "",
		FcmToken:     "",
		TOS:          time.Now().UTC().Format(time.RFC3339),
		Model:        "PC",
		SerialNumber: "",
		Locale:       systemLocale(),
	})
	if err != nil {
		return fmt.Errorf("registering device: %w", err)
	}

	acct = &config.Account{
		DeviceID:    resp.ID,
		AccessToken: resp.Token,
		PrivateKey:  privKey.String(),
	}
	if err := config.Save(ctx, acct); err != nil {
		slog.Error("failed to save config, manually save the following credentials",
			slog.String("device_id", resp.ID),
			slog.String("access_token", resp.Token),
			slog.String("private_key", privKey.String()),
		)
		return fmt.Errorf("saving config: %w", err)
	}

	slog.Info("registration successful",
		slog.String("device_id", resp.ID),
		slog.String("account_type", resp.Account.AccountType),
	)

	return nil
}

// systemLocale returns the system locale by checking LC_ALL, LC_MESSAGES, and
// LANG environment variables (in that order). The encoding suffix (e.g. ".UTF-8")
// is stripped. Falls back to "en_US" if no locale is set.
func systemLocale() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(key); v != "" {
			if loc, _, found := strings.Cut(v, "."); found {
				return loc
			}
			return v
		}
	}
	return "en_US"
}
