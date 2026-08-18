package model

import (
	"fmt"

	"razer-pair/internal/hid"
)

type Commands struct {
	ReceiverIdentity byte
	KeyboardPrepare  byte
	KeyboardCommit   byte
}

type Profile struct {
	Slug     string
	Name     string
	Keyboard hid.DeviceSpec
	Receiver hid.DeviceSpec
	Commands Commands
}

func (p Profile) Specs() []hid.DeviceSpec {
	return []hid.DeviceSpec{p.Keyboard, p.Receiver}
}

func (p Profile) Spec(role hid.Role) (hid.DeviceSpec, bool) {
	switch role {
	case hid.Keyboard:
		return p.Keyboard, true
	case hid.Receiver:
		return p.Receiver, true
	default:
		return hid.DeviceSpec{}, false
	}
}

var profiles = []Profile{
	{
		Slug: "pro-type-ultra",
		Name: "Razer Pro Type Ultra",
		Keyboard: hid.DeviceSpec{
			Role: hid.Keyboard, VendorID: 0x1532, ProductID: 0x0277,
			FeatureReportSize: 90, TransactionID: 0x1f,
		},
		Receiver: hid.DeviceSpec{
			Role: hid.Receiver, VendorID: 0x1532, ProductID: 0x027b,
			FeatureReportSize: 90, TransactionID: 0xff,
		},
		Commands: Commands{ReceiverIdentity: 0x95, KeyboardPrepare: 0x24, KeyboardCommit: 0xa4},
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
