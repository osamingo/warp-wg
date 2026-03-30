package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/peterbourgon/ff/v4"

	"github.com/osamingo/warp-wg/internal/config"
	"github.com/osamingo/warp-wg/internal/warp"
	"github.com/osamingo/warp-wg/internal/wireguard"
)

const defaultWARPPort = "2408"

func newProfileCmd() *ff.Command {
	flags := ff.NewFlagSet("profile")
	noIPv6 := flags.Bool(0, "no-ipv6", "Exclude IPv6 addresses and DNS")
	useIP := flags.Bool(0, "endpoint-ip", "Use IP address instead of hostname for endpoint")
	reserved := flags.Bool(0, "reserved", "Output Reserved field as active directive instead of comment")
	mtu := flags.IntLong("mtu", 1420, "MTU value (default: 1420, use 1280 for maximum compatibility)")
	port := flags.StringLong("port", defaultWARPPort, "Endpoint port (default: 2408)")

	return &ff.Command{
		Name:      "profile",
		Usage:     "warp-wg profile [--no-ipv6] [--endpoint-ip] [--reserved] [--mtu N] [--port N]",
		ShortHelp: "Output WireGuard profile to stdout",
		Flags:     flags,
		Exec: func(ctx context.Context, _ []string) error {
			if *mtu < 1280 || *mtu > 1500 {
				return fmt.Errorf("invalid MTU: %d (must be between 1280 and 1500)", *mtu)
			}
			if p, err := strconv.Atoi(*port); err != nil || p < 1 || p > 65535 {
				return fmt.Errorf("invalid port: %s (must be between 1 and 65535)", *port)
			}
			return execProfile(ctx, os.Stdout, profileFlags{
				noIPv6:   *noIPv6,
				useIP:    *useIP,
				reserved: *reserved,
				mtu:      *mtu,
				port:     *port,
			})
		},
	}
}

type profileFlags struct {
	noIPv6   bool
	useIP    bool
	reserved bool
	mtu      int
	port     string
}

func execProfile(ctx context.Context, out io.Writer, pf profileFlags) error {
	reg, err := config.LoadRegistered(ctx)
	if err != nil {
		return err
	}

	client := warp.NewClientFromContext(ctx)
	registration, err := client.Registration(ctx, reg.RegistrationID, reg.APIToken)
	if err != nil {
		return fmt.Errorf("fetching registration info: %w", err)
	}

	if len(registration.Config.Peers) == 0 {
		return fmt.Errorf("no peers found in registration config")
	}

	peer := registration.Config.Peers[0]

	var endpoint string
	if pf.useIP {
		endpoint = peerEndpoint(peer.Endpoint.V4, pf.port)
	} else {
		endpoint = peerEndpoint(peer.Endpoint.Host, pf.port)
	}

	reserved, err := wireguard.ClientIDToReserved(registration.Config.ClientID)
	if err != nil {
		return fmt.Errorf("converting client_id: %w", err)
	}

	data := wireguard.NewProfileData(
		reg.PrivateKey,
		registration.Config.Interface.Addresses.V4,
		registration.Config.Interface.Addresses.V6,
		peer.PublicKey,
		endpoint,
		reserved,
		wireguard.ProfileOptions{NoIPv6: pf.noIPv6, MTU: pf.mtu, Reserved: pf.reserved},
	)

	return wireguard.WriteProfile(out, data)
}

// peerEndpoint extracts the host from the API-provided endpoint address and
// replaces the port with the specified port. The API sometimes returns port 0
// or omits the port entirely, so we always override it.
func peerEndpoint(v4, port string) string {
	host, _, err := net.SplitHostPort(v4)
	if err != nil {
		host = v4
	}
	return net.JoinHostPort(host, port)
}
