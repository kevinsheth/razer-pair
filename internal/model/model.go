package model

import (
	"fmt"

	"razer-pair/internal/hid"
)

type Command struct {
	Target hid.Role
	ID     byte
}

type Commands struct {
	Identity Command
	Prepare  Command
	Commit   Command
}

type Profile struct {
	Slug       string
	Name       string
	Peripheral hid.DeviceSpec
	Receiver   hid.DeviceSpec
	Commands   Commands
}

func (p Profile) Specs() []hid.DeviceSpec {
	return []hid.DeviceSpec{p.Peripheral, p.Receiver}
}

func (p Profile) Spec(role hid.Role) (hid.DeviceSpec, bool) {
	switch role {
	case hid.Receiver:
		return p.Receiver, true
	default:
		if role == p.Peripheral.Role {
			return p.Peripheral, true
		}
		return hid.DeviceSpec{}, false
	}
}

var profiles = []Profile{
	{
		Slug: "pro-type-ultra",
		Name: "Razer Pro Type Ultra",
		Peripheral: hid.DeviceSpec{
			Role: hid.Peripheral, Label: "wired keyboard", VendorID: 0x1532, ProductID: 0x0277,
			FeatureReportSize: 90, TransactionID: 0x1f,
		},
		Receiver: hid.DeviceSpec{
			Role: hid.Receiver, Label: "receiver", VendorID: 0x1532, ProductID: 0x027b,
			FeatureReportSize: 90, TransactionID: 0xff,
		},
		Commands: Commands{
			Identity: Command{Target: hid.Receiver, ID: 0x95},
			Prepare:  Command{Target: hid.Peripheral, ID: 0x24},
			Commit:   Command{Target: hid.Peripheral, ID: 0xa4},
		},
	},
	{
		Slug: "basilisk-ultimate",
		Name: "Razer Basilisk Ultimate",
		Peripheral: hid.DeviceSpec{
			Role: hid.Peripheral, Label: "wired mouse", VendorID: 0x1532, ProductID: 0x0086,
			FeatureReportSize: 90, TransactionID: 0x1f,
		},
		Receiver: hid.DeviceSpec{
			Role: hid.Receiver, Label: "receiver", VendorID: 0x1532, ProductID: 0x0088,
			FeatureReportSize: 90, TransactionID: 0xff,
		},
		Commands: Commands{
			Identity: Command{Target: hid.Peripheral, ID: 0x97},
			Prepare:  Command{Target: hid.Receiver, ID: 0x15},
			Commit:   Command{Target: hid.Receiver, ID: 0x95},
		},
	},
}

func Get(slug string) (Profile, error) {
	for _, profile := range profiles {
		if profile.Slug == slug {
			return profile, nil
		}
	}
	return Profile{}, fmt.Errorf("unsupported model %q", slug)
}

func All() []Profile {
	return append([]Profile(nil), profiles...)
}

func Default() Profile {
	return profiles[0]
}
