package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"razer-pair/internal/hid"
	"razer-pair/internal/model"
	"razer-pair/internal/pairing"
)

func pair(ctx context.Context, provider hid.Provider, profile model.Profile, args []string, options Options) int {
	flags := flag.NewFlagSet("pair", flag.ContinueOnError)
	flags.SetOutput(options.Stderr)
	yes := flags.Bool("yes", false, "confirm pairing without an interactive prompt")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		if err == nil {
			fmt.Fprintln(options.Stderr, "error: pair accepts only --yes")
		}
		return ExitUsage
	}

	confirm := func() bool {
		fmt.Fprintf(options.Stdout, "Prepared %s; receiver identity received and device handshake validated.\n", profile.Name)
		if *yes {
			return true
		}
		fmt.Fprintf(options.Stdout, "Commit pairing to the %s? [y/N] ", profile.Peripheral.Label)
		return confirmed(options.Stdin)
	}

	err := pairing.Pair(ctx, provider, profile, confirm)
	if err != nil {
		if errors.Is(err, pairing.ErrCancelled) {
			fmt.Fprintln(options.Stderr, "Pairing cancelled; the commit command was not sent.")
			return ExitCancelled
		}
		fmt.Fprintln(options.Stderr, "error:", err)
		return ExitDevice
	}
	fmt.Fprintln(options.Stdout, "Pairing successful; post-commit receiver identity comparison passed.")
	return ExitOK
}

func confirmed(r io.Reader) bool {
	answer, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && answer == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
