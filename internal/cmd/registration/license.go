package registration

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
)

// NewLicenseCmd creates the "registration license" command.
func NewLicenseCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("license").SetParent(parentFlags)

	return &ff.Command{
		Name:      "license",
		Usage:     "warp-wg registration license <KEY>",
		ShortHelp: "Set a WARP+ license key",
		Flags:     flags,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: warp-wg registration license <KEY>")
			}
			return execLicense(ctx, args[0])
		},
	}
}

func execLicense(ctx context.Context, licenseKey string) error {
	acct, err := config.LoadRegistered(ctx)
	if err != nil {
		return err
	}

	client := warp.NewClientFromContext(ctx)
	resp, err := client.UpdateAccount(ctx, acct.DeviceID, acct.AccessToken, &warp.UpdateAccountRequest{
		License: licenseKey,
	})
	if err != nil {
		return fmt.Errorf("updating license: %w", err)
	}

	slog.Info("license updated",
		slog.String("account_type", resp.AccountType),
	)

	return nil
}
