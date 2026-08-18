//go:build cgo && (darwin || linux || windows || freebsd)

package hidapi

import (
	"fmt"

	"razer-pair/internal/protocol"
)

const wireReportSize = protocol.ReportSize + 1

func encodeFeatureReport(report protocol.Report) [wireReportSize]byte {
	var wire [wireReportSize]byte
	copy(wire[1:], report[:]) // HIDAPI reserves byte zero for the report ID.
	return wire
}

func decodeFeatureReport(wire [wireReportSize]byte, n int) (protocol.Report, error) {
	if n != len(wire) {
		return protocol.Report{}, fmt.Errorf("feature response length: got %d, want %d", n, len(wire))
	}
	if wire[0] != 0 {
		return protocol.Report{}, fmt.Errorf("feature response report ID: got 0x%02x, want 0x00", wire[0])
	}
	var report protocol.Report
	copy(report[:], wire[1:])
	return report, nil
}
