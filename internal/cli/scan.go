package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"razer-pair/internal/hid"
	"razer-pair/internal/model"
)

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

type deviceID struct {
	vendor  uint16
	product uint16
}

type featureInterface struct {
	number int
	size   int
}

func printScanSummary(w io.Writer, descriptors []hid.Descriptor) {
	groups := make(map[deviceID][]hid.Descriptor)
	for _, descriptor := range descriptors {
		key := deviceID{descriptor.VendorID, descriptor.ProductID}
		groups[key] = append(groups[key], descriptor)
	}
	ids := make([]deviceID, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].vendor < ids[j].vendor || ids[i].vendor == ids[j].vendor && ids[i].product < ids[j].product
	})
	for _, id := range ids {
		group := groups[id]
		fmt.Fprintf(w, "  %s interfaces=%s bus=%s product=%q%s\n",
			group[0].ID(), formatInterfaces(group), group[0].Transport, group[0].Product,
			formatProfileMatches(id))
	}
}

func formatProfileMatches(id deviceID) string {
	var matches []string
	for _, profile := range model.All() {
		for _, spec := range profile.Specs() {
			if spec.VendorID == id.vendor && spec.ProductID == id.product {
				matches = append(matches, profile.Slug+"/"+string(spec.Role))
			}
		}
	}
	if len(matches) == 0 {
		return ""
	}
	return " match=" + strings.Join(matches, ",")
}

func formatInterfaces(descriptors []hid.Descriptor) string {
	unique := make(map[featureInterface]bool)
	for _, descriptor := range descriptors {
		if descriptor.MaxFeatureReport > 0 {
			unique[featureInterface{descriptor.Interface, descriptor.MaxFeatureReport}] = true
		}
	}
	interfaces := make([]featureInterface, 0, len(unique))
	for candidate := range unique {
		interfaces = append(interfaces, candidate)
	}
	sort.Slice(interfaces, func(i, j int) bool {
		return interfaces[i].number < interfaces[j].number ||
			interfaces[i].number == interfaces[j].number && interfaces[i].size < interfaces[j].size
	})
	if len(interfaces) == 0 {
		return "unknown/inaccessible"
	}
	parts := make([]string, len(interfaces))
	for i, candidate := range interfaces {
		parts[i] = fmt.Sprintf("[%d feature=%d]", candidate.number, candidate.size)
	}
	return strings.Join(parts, " ")
}
