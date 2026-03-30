package registration

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
)

// NewUpdateCmd creates the "registration update" command.
func NewUpdateCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("update").SetParent(parentFlags)
	name := flags.StringLong("name", "", "Set device name")

	return &ff.Command{
		Name:      "update",
		Usage:     "warp-wg registration update --name <NAME>",
		ShortHelp: "Update device registration settings",
		Flags:     flags,
		Exec: func(ctx context.Context, _ []string) error {
			if *name == "" {
				return fmt.Errorf("at least one flag is required (e.g. --name)")
			}
			return execUpdate(ctx, *name)
		},
	}
}

func execUpdate(ctx context.Context, name string) error {
	reg, err := config.LoadRegistered(ctx)
	if err != nil {
		return err
	}

	client := warp.NewClientFromContext(ctx)
	if _, err := client.UpdateRegistrationKey(ctx, reg.RegistrationID, reg.APIToken, &warp.UpdateRegistrationRequest{
		Name: name,
	}); err != nil {
		return fmt.Errorf("updating registration: %w", err)
	}

	slog.Info("Registration updated", slog.String("name", name))

	return nil
}
