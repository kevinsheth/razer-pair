package mockhid

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"razer-pair/internal/hid"
	"razer-pair/internal/model"
)

type exchange struct {
	Command byte
	Input   [6]byte
	Output  [6]byte
	Err     error
}

type device struct {
	mu        sync.Mutex
	role      hid.Role
	exchanges []exchange
	closed    bool
}

func newDevice(role hid.Role, exchanges ...exchange) *device {
	return &device{role: role, exchanges: append([]exchange(nil), exchanges...)}
}

func (d *device) Command(ctx context.Context, command byte, input [6]byte) ([6]byte, error) {
	if err := ctx.Err(); err != nil {
		return [6]byte{}, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return [6]byte{}, errors.New("mock device is closed")
	}
	if len(d.exchanges) == 0 {
		return [6]byte{}, fmt.Errorf("unexpected %s command 0x%02x", d.role, command)
	}
	next := d.exchanges[0]
	d.exchanges = d.exchanges[1:]
	if command != next.Command {
		return [6]byte{}, fmt.Errorf("%s command: got 0x%02x, want 0x%02x", d.role, command, next.Command)
	}
	if input != next.Input {
		return [6]byte{}, fmt.Errorf("%s command 0x%02x input: got %x, want %x", d.role, command, input, next.Input)
	}
	return next.Output, next.Err
}

func (d *device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closed = true
	return nil
}

func (d *device) remaining() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.exchanges)
}

type Provider struct {
	descriptors []hid.Descriptor
	devices     map[hid.Role]*device
	openErrors  map[hid.Role]error
}

func (p *Provider) Scan(ctx context.Context, _ uint16) ([]hid.Descriptor, error) {
	return p.Enumerate(ctx, nil)
}

func (p *Provider) Enumerate(ctx context.Context, _ []hid.DeviceSpec) ([]hid.Descriptor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return append([]hid.Descriptor(nil), p.descriptors...), nil
}

func (p *Provider) Open(ctx context.Context, spec hid.DeviceSpec) (hid.Device, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := p.openErrors[spec.Role]; err != nil {
		return nil, err
	}
	device := p.devices[spec.Role]
	if device == nil {
		return nil, fmt.Errorf("%s %04x:%04x not found", spec.Role, spec.VendorID, spec.ProductID)
	}
	return device, nil
}

func (p *Provider) Remaining(role hid.Role) int {
	if device := p.devices[role]; device != nil {
		return device.remaining()
	}
	return 0
}

func Scenario(name string, profile model.Profile) (*Provider, error) {
	id := [6]byte{0x11, 0x22, 0x33, 0x44, 0x00, 0x00}
	provider := newProvider(profile)

	switch name {
	case "success":
		provider.addPairingDevices(profile, id, [6]byte{0x11, 0x22, 0x33, 0x44, 0x11, 0x22})
	case "mismatch":
		provider.addPairingDevices(profile, id, [6]byte{0xaa, 0xbb, 0xcc, 0xdd})
	case "missing-device":
		provider.descriptors = withoutRole(provider.descriptors, profile.Peripheral.Role)
		provider.addPairingDevices(profile, id, id)
		delete(provider.devices, profile.Peripheral.Role)
	case "denied":
		provider.openErrors[hid.Receiver] = errors.New("access denied (mock Input Monitoring failure)")
	default:
		return nil, fmt.Errorf("unknown mock scenario %q (choose success, mismatch, missing-device, or denied)", name)
	}
	return provider, nil
}

func newProvider(profile model.Profile) *Provider {
	descriptor := func(spec hid.DeviceSpec, product string) hid.Descriptor {
		return hid.Descriptor{
			Role: spec.Role, VendorID: spec.VendorID, ProductID: spec.ProductID,
			UsagePage: 0x01, Usage: 0x02, MaxFeatureReport: spec.FeatureReportSize,
			Transport: "USB", Product: product, Interface: 2,
		}
	}
	inaccessible := func(spec hid.DeviceSpec, product string) hid.Descriptor {
		descriptor := descriptor(spec, product)
		descriptor.Interface = 1
		descriptor.Usage = 0x06
		descriptor.MaxFeatureReport = 0
		descriptor.AccessError = "access denied (mock non-pairing collection)"
		return descriptor
	}
	return &Provider{
		descriptors: []hid.Descriptor{
			inaccessible(profile.Peripheral, profile.Name),
			descriptor(profile.Peripheral, profile.Name),
			inaccessible(profile.Receiver, profile.Name+" Dongle"),
			descriptor(profile.Receiver, profile.Name+" Dongle"),
		},
		devices:    make(map[hid.Role]*device),
		openErrors: make(map[hid.Role]error),
	}
}

func withoutRole(descriptors []hid.Descriptor, role hid.Role) []hid.Descriptor {
	filtered := make([]hid.Descriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.Role != role {
			filtered = append(filtered, descriptor)
		}
	}
	return filtered
}

func (p *Provider) addPairingDevices(profile model.Profile, receiverID, commit [6]byte) {
	var zero [6]byte
	p.devices[hid.Receiver] = newDevice(hid.Receiver,
		exchange{Command: profile.Commands.ReceiverIdentity, Input: zero, Output: receiverID},
	)
	p.devices[profile.Peripheral.Role] = newDevice(profile.Peripheral.Role,
		exchange{Command: profile.Commands.PeripheralPrepare, Input: receiverID, Output: receiverID},
		exchange{Command: profile.Commands.PeripheralCommit, Input: zero, Output: commit},
	)
}
