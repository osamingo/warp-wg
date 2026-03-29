package registration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
)

// NewDevicesCmd creates the "registration devices" command.
func NewDevicesCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("devices").SetParent(parentFlags)
	jsonOut := flags.Bool('j', "json", "Output as JSON")

	return &ff.Command{
		Name:      "devices",
		Usage:     "warp-wg registration devices [--json]",
		ShortHelp: "List devices linked to the account",
		Flags:     flags,
		Exec: func(ctx context.Context, _ []string) error {
			return execDevices(ctx, os.Stdout, *jsonOut)
		},
	}
}

func execDevices(ctx context.Context, out io.Writer, jsonOut bool) error {
	acct, err := config.LoadRegistered(ctx)
	if err != nil {
		return err
	}

	client := warp.NewClientFromContext(ctx)
	devices, err := client.BoundDevices(ctx, acct.DeviceID, acct.AccessToken)
	if err != nil {
		return fmt.Errorf("fetching devices: %w", err)
	}

	if jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(devices); err != nil {
			return fmt.Errorf("encoding json: %w", err)
		}
		return nil
	}

	if len(devices) == 0 {
		if _, err := fmt.Fprintln(out, "No devices found."); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		return nil
	}

	for _, d := range devices {
		active := "inactive"
		if d.Active {
			active = "active"
		}
		if _, err := fmt.Fprintf(out, "%-38s %-20s %-10s %s\n", d.ID, d.Name, active, d.Model); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	}

	return nil
}
