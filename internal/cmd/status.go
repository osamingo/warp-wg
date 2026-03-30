package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/peterbourgon/ff/v4"
)

const traceURL = "https://cloudflare.com/cdn-cgi/trace"

func newStatusCmd() *ff.Command {
	return &ff.Command{
		Name:      "status",
		Usage:     "warp-wg status",
		ShortHelp: "Show Cloudflare connection diagnostics",
		Exec: func(ctx context.Context, _ []string) error {
			return execStatus(ctx, os.Stdout, traceURL)
		},
	}
}

func execStatus(ctx context.Context, out io.Writer, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req) //nolint:gosec // url comes from a trusted constant or test
	if err != nil {
		return fmt.Errorf("fetching trace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("trace returned status %d", resp.StatusCode)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}
