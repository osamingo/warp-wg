package registration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
)

// NewShowCmd creates the "registration show" command.
func NewShowCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("show").SetParent(parentFlags)
	jsonOut := flags.Bool('j', "json", "Output as JSON")

	return &ff.Command{
		Name:      "show",
		Usage:     "warp-wg registration show [--json]",
		ShortHelp: "Show current registration details",
		Flags:     flags,
		Exec: func(ctx context.Context, _ []string) error {
			return execShow(ctx, os.Stdout, *jsonOut)
		},
	}
}

func execShow(ctx context.Context, out io.Writer, jsonOut bool) error {
	acct, err := config.LoadRegistered(ctx)
	if err != nil {
		return err
	}

	client := warp.NewClient()
	device, err := client.Device(ctx, acct.DeviceID, acct.AccessToken)
	if err != nil {
		return fmt.Errorf("fetching device info: %w", err)
	}

	if jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(device); err != nil {
			return fmt.Errorf("encoding json: %w", err)
		}
		return nil
	}

	return printDevice(out, device)
}

func printDevice(out io.Writer, d *warp.DeviceResponse) error {
	lines := []struct{ label, value string }{
		{"Device ID", d.ID},
		{"Account Type", d.Account.AccountType},
		{"Premium Data", humanize.IBytes(d.Account.PremiumData)},
		{"Quota", humanize.IBytes(d.Account.Quota)},
		{"Created", d.Account.Created},
		{"Updated", d.Account.Updated},
		{"IPv4", d.Config.Interface.Addresses.V4},
		{"IPv6", d.Config.Interface.Addresses.V6},
	}

	if len(d.Config.Peers) > 0 {
		p := d.Config.Peers[0]
		lines = append(lines,
			struct{ label, value string }{"Endpoint", p.Endpoint.Host},
			struct{ label, value string }{"Public Key", p.PublicKey},
		)
	}

	for _, l := range lines {
		if _, err := fmt.Fprintf(out, "%-14s%s\n", l.label+":", l.value); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	}

	return nil
}
