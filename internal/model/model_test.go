package model

import (
	"testing"

	"razer-pair/internal/hid"
)

func TestProTypeUltraProfile(t *testing.T) {
	profile := Default()
	if profile.Peripheral.VendorID != 0x1532 || profile.Peripheral.ProductID != 0x0277 ||
		profile.Receiver.VendorID != 0x1532 || profile.Receiver.ProductID != 0x027b {
		t.Fatalf("unexpected device IDs: %+v", profile)
	}
	if profile.Peripheral.FeatureReportSize != 90 || profile.Receiver.FeatureReportSize != 90 ||
		profile.Peripheral.TransactionID != 0x1f || profile.Receiver.TransactionID != 0xff {
		t.Fatalf("unexpected transport profile: %+v", profile)
	}
	if profile.Commands != (Commands{ReceiverIdentity: 0x95, PeripheralPrepare: 0x24, PeripheralCommit: 0xa4}) {
		t.Fatalf("unexpected commands: %+v", profile.Commands)
	}
}

func TestBasiliskUltimateProfile(t *testing.T) {
	profile, err := Get("basilisk-ultimate")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Peripheral.Role != hid.Mouse || profile.Peripheral.ProductID != 0x0086 || profile.Receiver.ProductID != 0x0088 {
		t.Fatalf("unexpected device profile: %+v", profile)
	}
	if profile.Commands != (Commands{ReceiverIdentity: 0x97, PeripheralPrepare: 0x15, PeripheralCommit: 0x95}) {
		t.Fatalf("unexpected commands: %+v", profile.Commands)
	}
}

func TestUnknownProfileIsRejected(t *testing.T) {
	if _, err := Get("looks-close-enough"); err == nil {
		t.Fatal("Get accepted an unknown profile")
	}
}
