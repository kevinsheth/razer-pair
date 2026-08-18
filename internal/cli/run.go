package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"razer-pair/internal/hid"
	"razer-pair/internal/mockhid"
	"razer-pair/internal/model"
	"razer-pair/internal/pairing"
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
	if command == "version" {
		if len(remaining) != 1 {
			fmt.Fprintln(options.Stderr, "error: version takes no arguments")
			return ExitUsage
		}
		fmt.Fprintf(options.Stdout, "razer-pair %s\n", options.Version)
		return ExitOK
	}
	if command == "list-models" {
		if len(remaining) != 1 {
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

	profile, err := model.Get(*modelSlug)
	if err != nil {
		fmt.Fprintln(options.Stderr, "error:", err)
		return ExitUsage
	}
	provider := options.RealProvider
	if *mockScenario != "" {
		provider, err = mockhid.Scenario(*mockScenario, profile)
		if err != nil {
			fmt.Fprintln(options.Stderr, "error:", err)
			return ExitUsage
		}
	}
	if provider == nil {
		fmt.Fprintln(options.Stderr, "error: no HID provider configured")
		return ExitFailure
	}

	switch command {
	case "scan":
		return scan(ctx, provider, remaining[1:], options)
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

func scan(ctx context.Context, provider hid.Provider, args []string, options Options) int {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(options.Stderr)
	verbose := flags.Bool("verbose", false, "show every HID collection and access error")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		if err == nil {
			fmt.Fprintln(options.Stderr, "error: scan accepts only --verbose")
		}
		return ExitUsage
	}
	descriptors, err := provider.Scan(ctx, 0x1532)
	if err != nil {
		fmt.Fprintln(options.Stderr, "error: scan devices:", err)
		return ExitFailure
	}
	if len(descriptors) == 0 {
		fmt.Fprintln(options.Stderr, "error: no Razer HID devices detected")
		return ExitDevice
	}
	fmt.Fprintln(options.Stdout, "Razer HID devices:")
	if *verbose {
		for _, descriptor := range descriptors {
			printCollection(options.Stdout, descriptor)
		}
	} else {
		printScanSummary(options.Stdout, descriptors)
	}
	fmt.Fprintln(options.Stdout, "No reports sent. Share this output when requesting device support.")
	return ExitOK
}

func printScanSummary(w io.Writer, descriptors []hid.Descriptor) {
	type id struct{ vendor, product uint16 }
	devices := make(map[id]hid.Descriptor)
	for _, descriptor := range descriptors {
		key := id{descriptor.VendorID, descriptor.ProductID}
		current, ok := devices[key]
		if !ok || scanRank(descriptor) > scanRank(current) {
			devices[key] = descriptor
		}
	}
	ids := make([]id, 0, len(devices))
	for key := range devices {
		ids = append(ids, key)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].vendor < ids[j].vendor || ids[i].vendor == ids[j].vendor && ids[i].product < ids[j].product
	})
	for _, key := range ids {
		descriptor := devices[key]
		feature := "unknown/inaccessible"
		if descriptor.MaxFeatureReport > 0 {
			feature = fmt.Sprint(descriptor.MaxFeatureReport)
		}
		fmt.Fprintf(w, "  %s feature=%s interface=%d bus=%s product=%q\n",
			descriptor.ID(), feature, descriptor.Interface, descriptor.Transport, descriptor.Product)
	}
}

func scanRank(descriptor hid.Descriptor) int {
	if descriptor.MaxFeatureReport == 90 {
		return 2
	}
	if descriptor.MaxFeatureReport > 0 {
		return 1
	}
	return 0
}

func inspectCommand(ctx context.Context, provider hid.Provider, profile model.Profile, args []string, options Options, dryRun bool) int {
	name := "inspect"
	if dryRun {
		name = "dry-run"
	}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(options.Stderr)
	verbose := flags.Bool("verbose", false, "show every HID collection and access error")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		if err == nil {
			fmt.Fprintf(options.Stderr, "error: %s accepts only --verbose\n", name)
		}
		return ExitUsage
	}
	return inspect(ctx, provider, profile, options.Stdout, options.Stderr, dryRun, *verbose)
}

func inspect(ctx context.Context, provider hid.Provider, profile model.Profile, stdout, stderr io.Writer, dryRun, verbose bool) int {
	descriptors, err := provider.Enumerate(ctx, profile.Specs())
	if err != nil {
		fmt.Fprintln(stderr, "error: inspect devices:", err)
		return ExitFailure
	}
	found := map[hid.Role]bool{}
	detected := map[hid.Role]bool{}
	printed := map[hid.Role]bool{}
	if verbose {
		fmt.Fprintf(stdout, "%s HID collections:\n", profile.Name)
	} else {
		fmt.Fprintf(stdout, "%s pairing interfaces:\n", profile.Name)
	}
	for _, descriptor := range descriptors {
		detected[descriptor.Role] = true
		spec, ok := profile.Spec(descriptor.Role)
		eligible := ok && matches(descriptor, spec)
		if eligible {
			found[descriptor.Role] = true
		}
		if verbose {
			printCollection(stdout, descriptor)
		} else if eligible {
			if !printed[descriptor.Role] {
				printVerified(stdout, descriptor)
				printed[descriptor.Role] = true
			}
		}
	}
	roles := []hid.Role{profile.Peripheral.Role, hid.Receiver}
	if !found[roles[0]] || !found[roles[1]] {
		printMissing(stderr, roles, found, detected)
		return ExitDevice
	}
	fmt.Fprintf(stdout, "Ready: exact %s and receiver feature interfaces found.\n", profile.Peripheral.Role.Label())
	if dryRun {
		fmt.Fprintf(stdout, "No reports sent. Planned sequence: receiver 0x%02x, device 0x%02x, confirmation, device 0x%02x.\n",
			profile.Commands.ReceiverIdentity, profile.Commands.PeripheralPrepare, profile.Commands.PeripheralCommit)
	}
	return ExitOK
}

func matches(descriptor hid.Descriptor, spec hid.DeviceSpec) bool {
	return descriptor.VendorID == spec.VendorID &&
		descriptor.ProductID == spec.ProductID &&
		descriptor.MaxFeatureReport == spec.FeatureReportSize
}

func printCollection(w io.Writer, descriptor hid.Descriptor) {
	feature := "unknown/inaccessible"
	if descriptor.MaxFeatureReport > 0 {
		feature = fmt.Sprint(descriptor.MaxFeatureReport)
	}
	if descriptor.Role == hid.Unknown {
		fmt.Fprintf(w, "  %s interface=%d usage=0x%04x:0x%04x feature=%s bus=%s product=%q\n",
			descriptor.ID(), descriptor.Interface, descriptor.UsagePage, descriptor.Usage,
			feature, descriptor.Transport, descriptor.Product)
	} else {
		fmt.Fprintf(w, "  %-14s %s interface=%d usage=0x%04x:0x%04x feature=%s bus=%s product=%q\n",
			descriptor.Role, descriptor.ID(), descriptor.Interface, descriptor.UsagePage,
			descriptor.Usage, feature, descriptor.Transport, descriptor.Product)
	}
	if descriptor.AccessError != "" {
		fmt.Fprintf(w, "    access: %s\n", descriptor.AccessError)
	}
}

func printVerified(w io.Writer, descriptor hid.Descriptor) {
	fmt.Fprintf(w, "  %-14s %s verified (interface=%d, feature=%d)\n",
		descriptor.Role, descriptor.ID(), descriptor.Interface, descriptor.MaxFeatureReport)
}

func printMissing(w io.Writer, roles []hid.Role, found, detected map[hid.Role]bool) {
	for _, role := range roles {
		if found[role] {
			continue
		}
		state := "not detected"
		if detected[role] {
			state = "detected, but its pairing interface is inaccessible or invalid"
		}
		fmt.Fprintf(w, "error: %s %s\n", role.Label(), state)
	}
	fmt.Fprintln(w, "Run again with --verbose for complete HID diagnostics.")
}

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
		fmt.Fprintf(options.Stdout, "Commit pairing to the %s? [y/N] ", profile.Peripheral.Role.Label())
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

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: razer-pair [--model MODEL] [--mock SCENARIO] <command> [options]")
	fmt.Fprintln(w, "Commands: scan [--verbose], inspect [--verbose], dry-run [--verbose], pair [--yes], list-models, version")
}
