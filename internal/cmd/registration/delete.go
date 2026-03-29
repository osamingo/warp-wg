package registration

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
)

// NewDeleteCmd creates the "registration delete" command.
func NewDeleteCmd(parentFlags *ff.FlagSet) *ff.Command {
	flags := ff.NewFlagSet("delete").SetParent(parentFlags)
	quiet := flags.Bool('q', "quiet", "Skip confirmation prompt")

	return &ff.Command{
		Name:      "delete",
		Usage:     "warp-wg registration delete [-q]",
		ShortHelp: "Delete current device registration",
		Flags:     flags,
		Exec: func(ctx context.Context, _ []string) error {
			if !*quiet {
				if err := confirmDelete(os.Stdin, os.Stderr); err != nil {
					return err
				}
			}
			return execDelete(ctx)
		},
	}
}

func confirmDelete(in io.Reader, out io.Writer) error {
	if _, err := fmt.Fprint(out, "This will delete the device registration. Continue? (y/N): "); err != nil {
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

func execDelete(ctx context.Context) error {
	acct, err := config.LoadRegistered(ctx)
	if err != nil {
		return err
	}

	slog.Info("deleting device registration", slog.String("device_id", acct.DeviceID))

	client := warp.NewClientFromContext(ctx)
	if err := client.DeleteDevice(ctx, acct.DeviceID, acct.AccessToken); err != nil {
		return fmt.Errorf("deleting device: %w", err)
	}

	cfgPath, err := config.FilePath(ctx)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	if err := os.Remove(cfgPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removing config file: %w", err)
	}

	os.Remove(filepath.Dir(cfgPath)) //nolint:errcheck,gosec // best-effort cleanup, directory may not be empty

	slog.Info("registration deleted")

	return nil
}
