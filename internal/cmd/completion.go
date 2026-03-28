package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/peterbourgon/ff/v4"
)

func newCompletionCmd() *ff.Command {
	return &ff.Command{
		Name:      "completion",
		Usage:     "warp-wg completion <bash|zsh|fish>",
		ShortHelp: "Generate shell completion script",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: warp-wg completion <bash|zsh|fish>")
			}
			return printCompletion(os.Stdout, args[0])
		},
	}
}

func printCompletion(out io.Writer, shell string) error {
	var script string

	switch shell {
	case "bash":
		script = bashCompletion
	case "zsh":
		script = zshCompletion
	case "fish":
		script = fishCompletion
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}

	if _, err := fmt.Fprint(out, script); err != nil {
		return fmt.Errorf("writing completion: %w", err)
	}

	return nil
}

const bashCompletion = `_warp_wg() {
    local cur prev words cword
    _init_completion || return

    local commands="registration profile status version completion"
    local reg_commands="new show delete license devices rotate-keys"

    case "${words[1]}" in
        registration)
            case "${words[2]}" in
                new)
                    COMPREPLY=($(compgen -W "--accept-tos --help" -- "$cur"))
                    return
                    ;;
                show)
                    COMPREPLY=($(compgen -W "--json -j --help" -- "$cur"))
                    return
                    ;;
                delete)
                    COMPREPLY=($(compgen -W "--quiet -q --help" -- "$cur"))
                    return
                    ;;
                devices)
                    COMPREPLY=($(compgen -W "--json -j --help" -- "$cur"))
                    return
                    ;;
                license|rotate-keys)
                    COMPREPLY=($(compgen -W "--help" -- "$cur"))
                    return
                    ;;
                *)
                    COMPREPLY=($(compgen -W "$reg_commands" -- "$cur"))
                    return
                    ;;
            esac
            ;;
        profile)
            COMPREPLY=($(compgen -W "--no-ipv6 --endpoint-ip --mtu --port --help" -- "$cur"))
            return
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur"))
            return
            ;;
        status|version)
            return
            ;;
        *)
            COMPREPLY=($(compgen -W "--config $commands" -- "$cur"))
            return
            ;;
    esac
}

complete -F _warp_wg warp-wg
`

const zshCompletion = `#compdef warp-wg

_warp_wg() {
    local -a commands
    commands=(
        'registration:Manage WARP device registration'
        'profile:Output WireGuard profile to stdout'
        'status:Show Cloudflare connection diagnostics'
        'version:Print version information'
        'completion:Generate shell completion script'
    )

    local -a reg_commands
    reg_commands=(
        'new:Register a new WARP device'
        'show:Show current registration details'
        'delete:Delete current device registration'
        'license:Set a WARP+ license key'
        'devices:List devices linked to the account'
        'rotate-keys:Generate a new key pair and update the registration'
    )

    _arguments -C \
        '--config[Path to config file]:path:_files' \
        '1:command:->command' \
        '*::arg:->args'

    case $state in
        command)
            _describe 'command' commands
            ;;
        args)
            case ${words[1]} in
                registration)
                    _arguments -C \
                        '1:subcommand:->subcmd' \
                        '*::arg:->subargs'
                    case $state in
                        subcmd)
                            _describe 'subcommand' reg_commands
                            ;;
                        subargs)
                            case ${words[1]} in
                                new)
                                    _arguments '--accept-tos[Accept the Cloudflare Terms of Service]'
                                    ;;
                                show)
                                    _arguments '(-j --json)'{-j,--json}'[Output as JSON]'
                                    ;;
                                delete)
                                    _arguments '(-q --quiet)'{-q,--quiet}'[Skip confirmation prompt]'
                                    ;;
                                devices)
                                    _arguments '(-j --json)'{-j,--json}'[Output as JSON]'
                                    ;;
                            esac
                            ;;
                    esac
                    ;;
                profile)
                    _arguments \
                        '--no-ipv6[Exclude IPv6 addresses and DNS]' \
                        '--endpoint-ip[Use IP address instead of hostname]' \
                        '--mtu[MTU value]:mtu:' \
                        '--port[Endpoint port]:port:'
                    ;;
                completion)
                    _arguments '1:shell:(bash zsh fish)'
                    ;;
            esac
            ;;
    esac
}

_warp_wg "$@"
`

const fishCompletion = `# Disable file completion by default
complete -c warp-wg -f

# Global flags
complete -c warp-wg -l config -d 'Path to config file' -r -F

# Top-level commands
complete -c warp-wg -n '__fish_use_subcommand' -a registration -d 'Manage WARP device registration'
complete -c warp-wg -n '__fish_use_subcommand' -a profile -d 'Output WireGuard profile to stdout'
complete -c warp-wg -n '__fish_use_subcommand' -a status -d 'Show Cloudflare connection diagnostics'
complete -c warp-wg -n '__fish_use_subcommand' -a version -d 'Print version information'
complete -c warp-wg -n '__fish_use_subcommand' -a completion -d 'Generate shell completion script'

# registration subcommands
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and not __fish_seen_subcommand_from new show delete license devices rotate-keys' -a new -d 'Register a new WARP device'
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and not __fish_seen_subcommand_from new show delete license devices rotate-keys' -a show -d 'Show current registration details'
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and not __fish_seen_subcommand_from new show delete license devices rotate-keys' -a delete -d 'Delete current device registration'
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and not __fish_seen_subcommand_from new show delete license devices rotate-keys' -a license -d 'Set a WARP+ license key'
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and not __fish_seen_subcommand_from new show delete license devices rotate-keys' -a devices -d 'List devices linked to the account'
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and not __fish_seen_subcommand_from new show delete license devices rotate-keys' -a rotate-keys -d 'Generate a new key pair and update the registration'

# registration new flags
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and __fish_seen_subcommand_from new' -l accept-tos -d 'Accept the Cloudflare Terms of Service'

# registration show flags
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and __fish_seen_subcommand_from show' -s j -l json -d 'Output as JSON'

# registration delete flags
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and __fish_seen_subcommand_from delete' -s q -l quiet -d 'Skip confirmation prompt'

# registration devices flags
complete -c warp-wg -n '__fish_seen_subcommand_from registration; and __fish_seen_subcommand_from devices' -s j -l json -d 'Output as JSON'

# profile flags
complete -c warp-wg -n '__fish_seen_subcommand_from profile' -l no-ipv6 -d 'Exclude IPv6 addresses and DNS'
complete -c warp-wg -n '__fish_seen_subcommand_from profile' -l endpoint-ip -d 'Use IP address instead of hostname'
complete -c warp-wg -n '__fish_seen_subcommand_from profile' -l mtu -d 'MTU value' -r
complete -c warp-wg -n '__fish_seen_subcommand_from profile' -l port -d 'Endpoint port' -r

# completion arguments
complete -c warp-wg -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish' -a 'bash zsh fish'
`
