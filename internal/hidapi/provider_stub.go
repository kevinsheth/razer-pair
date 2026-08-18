//go:build !cgo || (!darwin && !linux && !windows && !freebsd)

package hidapi

import (
	"context"
	"errors"

	"razer-pair/internal/hid"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Scan(context.Context, uint16) ([]hid.Descriptor, error) {
	return nil, errors.New("real HID access requires cgo on macOS, Linux, Windows, or FreeBSD; use --mock for hardware-free testing")
}

func (p *Provider) Enumerate(context.Context, []hid.DeviceSpec) ([]hid.Descriptor, error) {
	return nil, errors.New("real HID access requires cgo on macOS, Linux, Windows, or FreeBSD; use --mock for hardware-free testing")
}

func (p *Provider) Open(context.Context, hid.DeviceSpec) (hid.Device, error) {
	return nil, errors.New("real HID access requires cgo on macOS, Linux, Windows, or FreeBSD; use --mock for hardware-free testing")
}
