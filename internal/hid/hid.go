package hid

import (
	"context"
	"fmt"
)

type Role string

const (
	Keyboard Role = "wired-keyboard"
	Receiver Role = "receiver"
)

type DeviceSpec struct {
	Role              Role
	VendorID          uint16
	ProductID         uint16
	FeatureReportSize int
	TransactionID     byte
}

type Descriptor struct {
	Role             Role
	VendorID         uint16
	ProductID        uint16
	UsagePage        uint32
	Usage            uint32
	MaxFeatureReport int
	Transport        string
	Product          string
	Interface        int
	AccessError      string
}

func (d Descriptor) ID() string {
	return fmt.Sprintf("%04x:%04x", d.VendorID, d.ProductID)
}

type Device interface {
	Command(ctx context.Context, command byte, input [6]byte) ([6]byte, error)
	Close() error
}

type Provider interface {
	Enumerate(ctx context.Context, specs []DeviceSpec) ([]Descriptor, error)
	Open(ctx context.Context, spec DeviceSpec) (Device, error)
}
