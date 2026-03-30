package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	applog "github.com/osamingo/warp-wg/internal/log"
)

// captureStderr replaces os.Stderr with a pipe, calls fn, then returns the
// captured output. The original stderr is restored after fn returns.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	origStderr := os.Stderr
	os.Stderr = w

	fn()

	os.Stderr = origStderr
	w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}

	return buf.String()
}

func TestHandler_Handle(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	tests := []struct {
		name  string
		level slog.Level
		msg   string
		attrs []slog.Attr
		want  string
	}{
		{
			name:  "success: info message without attrs",
			level: slog.LevelInfo,
			msg:   "Registering with Cloudflare WARP...",
			want:  "Registering with Cloudflare WARP...\n",
		},
		{
			name:  "success: info message with attrs",
			level: slog.LevelInfo,
			msg:   "Registration successful",
			attrs: []slog.Attr{
				slog.String("registration_id", "abc-123"),
				slog.String("account_type", "free"),
			},
			want: "Registration successful\n" +
				"  registration_id = abc-123\n" +
				"  account_type = free\n",
		},
		{
			name:  "success: warn message",
			level: slog.LevelWarn,
			msg:   "Something might be wrong",
			want:  "[WARNING] Something might be wrong\n",
		},
		{
			name:  "success: error message without attrs",
			level: slog.LevelError,
			msg:   "connection refused",
			want:  "[ERROR] connection refused\n",
		},
		{
			name:  "success: error message with attrs",
			level: slog.LevelError,
			msg:   "Failed to save config",
			attrs: []slog.Attr{
				slog.String("registration_id", "abc-123"),
				slog.String("private_key", "secret"),
			},
			want: "[ERROR] Failed to save config\n" +
				"  registration_id = abc-123\n" +
				"  private_key = secret\n",
		},
	}

	h := applog.NewHandler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := slog.NewRecord(time.Time{}, tt.level, tt.msg, 0)
			record.AddAttrs(tt.attrs...)

			got := captureStderr(t, func() {
				if err := h.Handle(context.Background(), record); err != nil {
					t.Fatalf("Handle() error = %v", err)
				}
			})

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Handle() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHandler_Enabled(t *testing.T) {
	h := applog.NewHandler()

	levels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	for _, level := range levels {
		if !h.Enabled(context.Background(), level) {
			t.Errorf("Enabled(%v) = false, want true", level)
		}
	}
}
