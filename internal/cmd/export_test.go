package cmd

import (
	"context"
	"io"
)

// Export unexported functions for testing.
var (
	PrintCompletion = printCompletion
	ExecStatus      = execStatus
	PeerEndpoint    = peerEndpoint
)

// ProfileFlags exposes profileFlags for testing.
type ProfileFlags struct {
	NoIPv6   bool
	UseIP    bool
	Reserved bool
	MTU      int
	Port     string
}

// ExecProfile wraps execProfile for testing.
func ExecProfile(ctx context.Context, out io.Writer, pf ProfileFlags) error {
	return execProfile(ctx, out, profileFlags{
		noIPv6:   pf.NoIPv6,
		useIP:    pf.UseIP,
		reserved: pf.Reserved,
		mtu:      pf.MTU,
		port:     pf.Port,
	})
}
