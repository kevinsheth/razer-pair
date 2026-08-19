package model

import (
	"testing"

	"razer-pair/internal/hid"
)

func TestProfiles(t *testing.T) {
	tests := []struct {
		slug                     string
		peripheralID, receiverID uint16
		commands                 Commands
	}{
		{
			slug: "pro-type-ultra", peripheralID: 0x0277, receiverID: 0x027b,
			commands: Commands{
				Identity: Command{Target: hid.Receiver, ID: 0x95},
				Prepare:  Command{Target: hid.Peripheral, ID: 0x24},
				Commit:   Command{Target: hid.Peripheral, ID: 0xa4},
			},
		},
		{
			slug: "basilisk-ultimate", peripheralID: 0x0086, receiverID: 0x0088,
			commands: Commands{
				Identity: Command{Target: hid.Peripheral, ID: 0x97},
				Prepare:  Command{Target: hid.Receiver, ID: 0x15},
				Commit:   Command{Target: hid.Receiver, ID: 0x95},
			},
		},
		{
			slug: "pro-click-v2", peripheralID: 0x00d0, receiverID: 0x00d1,
			commands: Commands{
				Identity: Command{Target: hid.Receiver, ID: 0x95},
				Prepare:  Command{Target: hid.Peripheral, ID: 0x24},
				Commit:   Command{Target: hid.Peripheral, ID: 0xa4},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.slug, func(t *testing.T) {
			profile, err := Get(test.slug)
			if err != nil {
				t.Fatal(err)
			}
			if profile.Peripheral.VendorID != 0x1532 || profile.Peripheral.ProductID != test.peripheralID ||
				profile.Receiver.VendorID != 0x1532 || profile.Receiver.ProductID != test.receiverID {
				t.Fatalf("unexpected device IDs: %+v", profile)
			}
			if profile.Peripheral.Role != hid.Peripheral || profile.Receiver.Role != hid.Receiver ||
				profile.Peripheral.FeatureReportSize != 90 || profile.Receiver.FeatureReportSize != 90 ||
				profile.Peripheral.TransactionID != 0x1f || profile.Receiver.TransactionID != 0xff {
				t.Fatalf("unexpected transport profile: %+v", profile)
			}
			if profile.Commands != test.commands {
				t.Fatalf("commands = %+v, want %+v", profile.Commands, test.commands)
			}
		})
	}
}

func TestDefaultProfile(t *testing.T) {
	if got := Default().Slug; got != "pro-type-ultra" {
		t.Fatalf("Default().Slug = %q", got)
	}
}

func TestUnknownProfileIsRejected(t *testing.T) {
	if _, err := Get("looks-close-enough"); err == nil {
		t.Fatal("Get accepted an unknown profile")
	}
}
