//go:build cgo && (darwin || linux || windows || freebsd)

package hidapi

import "testing"

func TestMaxFeatureReportSize(t *testing.T) {
	tests := []struct {
		name       string
		descriptor []byte
		want       int
	}{
		{name: "unnumbered 90 bytes", descriptor: []byte{0x75, 0x08, 0x95, 0x5a, 0xb1, 0x02}, want: 90},
		{name: "numbered 89 plus report ID", descriptor: []byte{0x85, 0x01, 0x75, 0x08, 0x95, 0x59, 0xb1, 0x02}, want: 90},
		{name: "largest of two reports", descriptor: []byte{0x85, 0x01, 0x75, 0x08, 0x95, 0x10, 0xb1, 0x02, 0x85, 0x02, 0x95, 0x59, 0xb1, 0x02}, want: 90},
		{name: "truncated", descriptor: []byte{0x75}, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := maxFeatureReportSize(test.descriptor); got != test.want {
				t.Fatalf("maxFeatureReportSize() = %d, want %d", got, test.want)
			}
		})
	}
}
