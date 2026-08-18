//go:build cgo && (darwin || linux || windows || freebsd)

package hidapi

import (
	"testing"

	"razer-pair/internal/protocol"
)

func TestFeatureReportWireFraming(t *testing.T) {
	report := protocol.BuildRequest(0xff, 0x95, [6]byte{})
	wire := encodeFeatureReport(report)
	if wire[0] != 0 {
		t.Fatalf("report ID = 0x%02x, want zero", wire[0])
	}
	if got := protocol.Report(wire[1:]); got != report {
		t.Fatal("wire payload differs from protocol report")
	}
	decoded, err := decodeFeatureReport(wire, len(wire))
	if err != nil {
		t.Fatal(err)
	}
	if decoded != report {
		t.Fatal("decoded report differs from original")
	}
}

func TestDecodeFeatureReportRejectsBadFrame(t *testing.T) {
	var wire [wireReportSize]byte
	if _, err := decodeFeatureReport(wire, protocol.ReportSize); err == nil {
		t.Fatal("accepted response without report-ID byte")
	}
	wire[0] = 1
	if _, err := decodeFeatureReport(wire, len(wire)); err == nil {
		t.Fatal("accepted unexpected report ID")
	}
}
