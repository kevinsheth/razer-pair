package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"razer-pair/internal/hid"
	"razer-pair/internal/model"
)

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
		} else if eligible && !printed[descriptor.Role] {
			printVerified(stdout, descriptor)
			printed[descriptor.Role] = true
		}
	}
	specs := profile.Specs()
	for _, spec := range specs {
		if !found[spec.Role] {
			printMissing(stderr, specs, found, detected)
			return ExitDevice
		}
	}
	fmt.Fprintf(stdout, "Ready: exact %s and receiver feature interfaces found.\n", profile.Peripheral.Label)
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
			descriptor.RoleLabel(), descriptor.ID(), descriptor.Interface, descriptor.UsagePage,
			descriptor.Usage, feature, descriptor.Transport, descriptor.Product)
	}
	if descriptor.AccessError != "" {
		fmt.Fprintf(w, "    access: %s\n", descriptor.AccessError)
	}
}

func printVerified(w io.Writer, descriptor hid.Descriptor) {
	fmt.Fprintf(w, "  %-14s %s verified (interface=%d, feature=%d)\n",
		descriptor.RoleLabel(), descriptor.ID(), descriptor.Interface, descriptor.MaxFeatureReport)
}

func printMissing(w io.Writer, specs []hid.DeviceSpec, found, detected map[hid.Role]bool) {
	for _, spec := range specs {
		if found[spec.Role] {
			continue
		}
		state := "not detected"
		if detected[spec.Role] {
			state = "detected, but its pairing interface is inaccessible or invalid"
		}
		fmt.Fprintf(w, "error: %s %s\n", spec.Label, state)
	}
	fmt.Fprintln(w, "Run again with --verbose for complete HID diagnostics.")
}
