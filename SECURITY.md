# Security Policy

## Disclaimer

This is an **unofficial** tool that interacts with undocumented Cloudflare WARP APIs. It is not affiliated with, endorsed by, or supported by Cloudflare, Inc. Use at your own risk.

This project makes no guarantees about the availability, reliability, or security of the underlying Cloudflare WARP API. The API may change or be discontinued without notice.

## Reporting a Vulnerability

**Please do NOT report security vulnerabilities through public GitHub issues.**

Use [GitHub's private vulnerability reporting](https://github.com/osamingo/warp-wg/security/advisories/new) instead.

Please include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact

## Scope

This project's security scope is limited to:

- The `warp-wg` CLI and its source code
- Private key generation and handling
- Local configuration file (`reg.json`) management

Out of scope:

- Cloudflare WARP API or infrastructure
- WireGuard protocol

## Response

This project is maintained by a single developer. Best-effort response, but no guaranteed timelines.
