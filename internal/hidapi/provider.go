//go:build cgo && (darwin || linux || windows || freebsd)

package hidapi

import (
	"context"
	"fmt"
	"sync"

	gohid "github.com/sstallion/go-hid"

	"razer-pair/internal/hid"
)

var (
	initOnce sync.Once
	initErr  error
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func initialize() error {
	initOnce.Do(func() { initErr = gohid.Init() })
	return initErr
}

func (p *Provider) Scan(ctx context.Context, vendorID uint16) ([]hid.Descriptor, error) {
	if err := initialize(); err != nil {
		return nil, fmt.Errorf("initialize HIDAPI: %w", err)
	}
	return enumerate(ctx, vendorID, gohid.ProductIDAny, hid.Unknown, make(map[string]reportInfo))
}

func (p *Provider) Enumerate(ctx context.Context, specs []hid.DeviceSpec) ([]hid.Descriptor, error) {
	if err := initialize(); err != nil {
		return nil, fmt.Errorf("initialize HIDAPI: %w", err)
	}
	var descriptors []hid.Descriptor
	reports := make(map[string]reportInfo)
	for _, spec := range specs {
		found, err := enumerate(ctx, spec.VendorID, spec.ProductID, spec.Role, reports)
		if err != nil {
			return nil, fmt.Errorf("enumerate %s: %w", spec.Role, err)
		}
		descriptors = append(descriptors, found...)
	}
	return descriptors, nil
}

func enumerate(ctx context.Context, vendorID, productID uint16, role hid.Role, reports map[string]reportInfo) ([]hid.Descriptor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var descriptors []hid.Descriptor
	err := gohid.Enumerate(vendorID, productID, func(info *gohid.DeviceInfo) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		report, ok := reports[info.Path]
		if !ok {
			report.size, report.err = reportSizeForPath(info.Path)
			reports[info.Path] = report
		}
		descriptor := hid.Descriptor{
			Role: role, VendorID: info.VendorID, ProductID: info.ProductID,
			UsagePage: uint32(info.UsagePage), Usage: uint32(info.Usage),
			MaxFeatureReport: report.size, Transport: info.BusType.String(),
			Product: info.ProductStr, Interface: info.InterfaceNbr,
		}
		if report.err != nil {
			descriptor.AccessError = report.err.Error()
		}
		descriptors = append(descriptors, descriptor)
		return nil
	})
	return descriptors, err
}

type reportInfo struct {
	size int
	err  error
}

func reportSizeForPath(path string) (int, error) {
	device, err := gohid.OpenPath(path)
	if err != nil {
		return 0, err
	}
	defer device.Close()
	return featureReportSize(device)
}

func (p *Provider) Open(ctx context.Context, spec hid.DeviceSpec) (hid.Device, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := initialize(); err != nil {
		return nil, fmt.Errorf("initialize HIDAPI: %w", err)
	}
	var selected *gohid.Device
	var inspected int
	var accessErr error
	seen := make(map[string]bool)
	err := gohid.Enumerate(spec.VendorID, spec.ProductID, func(info *gohid.DeviceInfo) error {
		if selected != nil || seen[info.Path] {
			return nil
		}
		seen[info.Path] = true
		if err := ctx.Err(); err != nil {
			return err
		}
		inspected++
		device, err := openFeatureDevice(info.Path, spec.FeatureReportSize)
		if err != nil {
			accessErr = err
			return nil
		}
		selected = device
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("enumerate %s: %w", spec.Role, err)
	}
	if selected == nil {
		if inspected == 0 {
			return nil, fmt.Errorf("%s %04x:%04x not found", spec.Role, spec.VendorID, spec.ProductID)
		}
		if accessErr != nil {
			return nil, fmt.Errorf("no accessible %d-byte feature interface for %s: %w", spec.FeatureReportSize, spec.Role, accessErr)
		}
		return nil, fmt.Errorf("%s has no %d-byte feature interface; refusing to send commands", spec.Role, spec.FeatureReportSize)
	}
	return &device{handle: selected, transactionID: spec.TransactionID}, nil
}

func openFeatureDevice(path string, wantSize int) (*gohid.Device, error) {
	device, err := gohid.OpenPath(path)
	if err != nil {
		return nil, err
	}
	size, err := featureReportSize(device)
	if err == nil && size != wantSize {
		err = fmt.Errorf("feature report size is %d, want %d", size, wantSize)
	}
	if err != nil {
		device.Close()
		return nil, err
	}
	return device, nil
}

func featureReportSize(device *gohid.Device) (int, error) {
	descriptor := make([]byte, 4096)
	n, err := device.GetReportDescriptor(descriptor)
	if err != nil {
		return 0, err
	}
	if n < 0 || n > len(descriptor) {
		return 0, fmt.Errorf("invalid report descriptor length %d", n)
	}
	return maxFeatureReportSize(descriptor[:n]), nil
}
