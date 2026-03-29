# warp-wg

[![CI](https://github.com/osamingo/warp-wg/actions/workflows/ci.yml/badge.svg)](https://github.com/osamingo/warp-wg/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/osamingo/warp-wg.svg)](https://pkg.go.dev/github.com/osamingo/warp-wg)
[![Go Report Card](https://goreportcard.com/badge/github.com/osamingo/warp-wg)](https://goreportcard.com/report/github.com/osamingo/warp-wg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Unofficial CLI for generating WireGuard profiles from [Cloudflare WARP](https://1.1.1.1/).

## Installation

### Go

```bash
go install github.com/osamingo/warp-wg/cmd/warp-wg@latest
```

### Binary

Download from [GitHub Releases](https://github.com/osamingo/warp-wg/releases) for all platforms.

## Usage

```bash
# Register a new device
warp-wg registration new

# Output WireGuard profile
warp-wg profile > warp.conf

# Show registration details (optional)
warp-wg registration show
```

## Commands

```
warp-wg
├── registration
│   ├── new           Register a new WARP device
│   ├── show          Show current registration details
│   ├── update        Update device registration settings
│   ├── delete        Delete current device registration
│   ├── license       Set a WARP+ license key
│   ├── devices       List devices linked to the account
│   └── rotate-keys   Generate a new key pair and update the registration
├── profile           Output WireGuard profile to stdout
├── status            Show Cloudflare connection diagnostics
├── completion        Generate shell completion script
└── version           Print version information
```

## Shell Completion

```bash
# bash (add to ~/.bashrc)
eval "$(warp-wg completion bash)"

# zsh (add to ~/.zshrc)
eval "$(warp-wg completion zsh)"

# fish
warp-wg completion fish | source
warp-wg completion fish > ~/.config/fish/completions/warp-wg.fish  # persistent
```

## Configuration

Credentials are stored in `~/.config/warp-wg/reg.json` ([XDG Base Directory](https://specifications.freedesktop.org/basedir-spec/latest/)) with `0600` permissions. The WireGuard private key is generated locally and never sent to the server.

Environment variable overrides:

| Variable | Description |
|----------|-------------|
| `WARP_WG_REGISTRATION_ID` | Registration ID |
| `WARP_WG_API_TOKEN` | API token |
| `WARP_WG_PRIVATE_KEY` | WireGuard private key |

## WARP+

To use a WARP+ license key, bind it after registration:

```bash
warp-wg registration license <KEY>
```

- Only keys purchased from the official [1.1.1.1](https://1.1.1.1/) app are supported.
- Up to 5 devices can be linked to a single account.

## Troubleshooting

If the standard WireGuard client fails to connect (handshake succeeds but no data flows), Cloudflare may be blocking connections without the correct `Reserved` bytes. The generated profile includes the `Reserved` value as a comment:

```ini
# Reserved = 171, 85, 205
```

Use this value with clients that support the `reserved` field, such as [Xray-core](https://github.com/XTLS/Xray-core) or [sing-box](https://github.com/SagerNet/sing-box).

## Disclaimer

> This project is not affiliated, associated, authorized, endorsed by, or in any way officially connected with Cloudflare, Inc.

- This tool uses an undocumented Cloudflare API. There is no stability guarantee; Cloudflare may change or remove the API at any time.
- Use of this tool may be subject to the [Cloudflare Terms of Service](https://www.cloudflare.com/application/terms/).
- Cloudflare is migrating from WireGuard to [MASQUE](https://blog.cloudflare.com/masque-now-powers-1-1-1-1-and-warp-apps-dex-available-with-remote-captures/). WireGuard-based connections may stop working in the future.

## License

[MIT](LICENSE)
