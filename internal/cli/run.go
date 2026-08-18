package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"razer-pair/internal/hid"
	"razer-pair/internal/mockhid"
	"razer-pair/internal/model"
)

const (
	ExitOK        = 0
	ExitFailure   = 1
	ExitDevice    = 2
	ExitCancelled = 3
	ExitUsage     = 64
)

type Options struct {
	Version      string
	RealProvider hid.Provider
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
}

func Run(ctx context.Context, args []string, options Options) int {
	global := flag.NewFlagSet("razer-pair", flag.ContinueOnError)
	global.SetOutput(options.Stderr)
	modelSlug := global.String("model", "pro-type-ultra", "device model profile")
	mockScenario := global.String("mock", "", "use a hardware-free mock scenario")
	global.Usage = func() { writeUsage(options.Stderr) }
	if err := global.Parse(args); err != nil {
		return ExitUsage
	}
	remaining := global.Args()
	if len(remaining) == 0 {
		writeUsage(options.Stderr)
		return ExitUsage
	}

	command := remaining[0]
	switch command {
	case "version":
		if len(remaining) != 1 {
			fmt.Fprintln(options.Stderr, "error: version takes no arguments")
			return ExitUsage
		}
		fmt.Fprintf(options.Stdout, "razer-pair %s\n", options.Version)
		return ExitOK
	case "list-models":
		return listModels(remaining[1:], options)
	case "scan":
		provider, err := selectProvider(options.RealProvider, *mockScenario, model.Default())
		if err != nil {
			return printProviderError(options.Stderr, err, *mockScenario)
		}
		return scan(ctx, provider, remaining[1:], options)
	}

	profile, err := model.Get(*modelSlug)
	if err != nil {
		fmt.Fprintln(options.Stderr, "error:", err)
		return ExitUsage
	}
	provider, err := selectProvider(options.RealProvider, *mockScenario, profile)
	if err != nil {
		return printProviderError(options.Stderr, err, *mockScenario)
	}
	switch command {
	case "inspect":
		return inspectCommand(ctx, provider, profile, remaining[1:], options, false)
	case "dry-run":
		return inspectCommand(ctx, provider, profile, remaining[1:], options, true)
	case "pair":
		return pair(ctx, provider, profile, remaining[1:], options)
	default:
		fmt.Fprintf(options.Stderr, "error: unknown command %q\n", command)
		writeUsage(options.Stderr)
		return ExitUsage
	}
}

func selectProvider(real hid.Provider, scenario string, profile model.Profile) (hid.Provider, error) {
	if scenario != "" {
		return mockhid.Scenario(scenario, profile)
	}
	if real == nil {
		return nil, errors.New("no HID provider configured")
	}
	return real, nil
}

func printProviderError(w io.Writer, err error, scenario string) int {
	fmt.Fprintln(w, "error:", err)
	if scenario != "" {
		return ExitUsage
	}
	return ExitFailure
}

func listModels(args []string, options Options) int {
	if len(args) != 0 {
		fmt.Fprintln(options.Stderr, "error: list-models takes no arguments")
		return ExitUsage
	}
	for _, profile := range model.All() {
		fmt.Fprintf(options.Stdout, "%s\t%s\tdevice=%04x:%04x receiver=%04x:%04x\n",
			profile.Slug, profile.Name,
			profile.Peripheral.VendorID, profile.Peripheral.ProductID,
			profile.Receiver.VendorID, profile.Receiver.ProductID)
	}
	return ExitOK
}

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: razer-pair [--model MODEL] [--mock SCENARIO] <command> [options]")
	fmt.Fprintln(w, "Commands: scan [--verbose], inspect [--verbose], dry-run [--verbose], pair [--yes], list-models, version")
}
