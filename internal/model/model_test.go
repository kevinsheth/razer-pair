package model

import "testing"

func TestProTypeUltraProfile(t *testing.T) {
	profile := Default()
	if profile.Keyboard.VendorID != 0x1532 || profile.Keyboard.ProductID != 0x0277 ||
		profile.Receiver.VendorID != 0x1532 || profile.Receiver.ProductID != 0x027b {
		t.Fatalf("unexpected device IDs: %+v", profile)
	}
	if profile.Keyboard.FeatureReportSize != 90 || profile.Receiver.FeatureReportSize != 90 ||
		profile.Keyboard.TransactionID != 0x1f || profile.Receiver.TransactionID != 0xff {
		t.Fatalf("unexpected transport profile: %+v", profile)
	}
	if profile.Commands != (Commands{ReceiverIdentity: 0x95, KeyboardPrepare: 0x24, KeyboardCommit: 0xa4}) {
		t.Fatalf("unexpected commands: %+v", profile.Commands)
	}
}

func TestUnknownProfileIsRejected(t *testing.T) {
	if _, err := Get("looks-close-enough"); err == nil {
		t.Fatal("Get accepted an unknown profile")
	}
}
