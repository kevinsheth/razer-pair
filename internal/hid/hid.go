package hid

import (
	"context"
	"fmt"
)

type Role string

const (
	Keyboard Role = "wired-keyboard"
	Mouse    Role = "wired-mouse"
	Receiver Role = "receiver"
	Unknown  Role = "unclassified"
)

func (r Role) Label() string {
	switch r {
	case Keyboard:
		return "wired keyboard"
	case Mouse:
		return "wired mouse"
	case Receiver:
		return "receiver"
	default:
		return "device"
	}
}

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
	Scan(context.Context, uint16) ([]Descriptor, error)
	Enumerate(ctx context.Context, specs []DeviceSpec) ([]Descriptor, error)
	Open(ctx context.Context, spec DeviceSpec) (Device, error)
}
