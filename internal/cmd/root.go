package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"

	"github.com/osamingo/warp-wg/internal/cmd/registration"
	"github.com/osamingo/warp-wg/internal/config"
)

var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		version = info.Main.Version
	}
}

func Run(ctx context.Context, args []string) error {
	rootFlags := ff.NewFlagSet("warp-wg")
	configPath := rootFlags.StringLong("config", "", "Path to config file")

	rootCmd := &ff.Command{
		Name:      "warp-wg",
		Usage:     "warp-wg [--config PATH] <subcommand> [FLAGS]",
		ShortHelp: "Unofficial WireGuard profile generator for Cloudflare WARP",
		Flags:     rootFlags,
		Subcommands: []*ff.Command{
			newRegistrationCmd(rootFlags),
			newProfileCmd(),
			newStatusCmd(),
			newCompletionCmd(),
			newVersionCmd(),
		},
	}

	if err := rootCmd.Parse(args); err != nil {
		if errors.Is(err, ff.ErrHelp) || errors.Is(err, ff.ErrNoExec) {
			fmt.Fprint(os.Stderr, ffhelp.Command(rootCmd))
			return nil
		}
		return err
	}

	if *configPath != "" {
		ctx = config.WithPath(ctx, *configPath)
	}

	err := rootCmd.Run(ctx)
	if errors.Is(err, ff.ErrNoExec) {
		fmt.Fprint(os.Stderr, ffhelp.Command(rootCmd))
		return nil
	}
	return err
}

func newRegistrationCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("registration").SetParent(parentFlags)

	return &ff.Command{
		Name:      "registration",
		Usage:     "warp-wg registration <subcommand> [FLAGS]",
		ShortHelp: "Manage WARP device registration",
		Flags:     flags,
		Subcommands: []*ff.Command{
			registration.NewNewCmd(flags),
			registration.NewShowCmd(flags),
			registration.NewUpdateCmd(flags),
			registration.NewDeleteCmd(flags),
			registration.NewLicenseCmd(flags),
			registration.NewDevicesCmd(flags),
			registration.NewRotateKeysCmd(flags),
		},
	}
}

func newVersionCmd() *ff.Command {
	return &ff.Command{
		Name:      "version",
		Usage:     "warp-wg version",
		ShortHelp: "Print version information",
		Exec: func(_ context.Context, _ []string) error {
			fmt.Println(version)
			return nil
		},
	}
}
