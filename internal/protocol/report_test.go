package protocol

import "testing"

func TestBuildRequestGolden(t *testing.T) {
	input := [6]byte{0x11, 0x22, 0x33, 0x44, 0, 0}
	var want Report
	want[1], want[5], want[7], want[88] = 0x1f, 6, 0x24, 0x66
	copy(want[8:], input[:])

	if got := BuildRequest(0x1f, 0x24, input); got != want {
		t.Fatalf("request:\n got %x\nwant %x", got, want)
	}
}

func TestParseResponse(t *testing.T) {
	want := [6]byte{0x11, 0x22, 0x33, 0x44, 0x11, 0x22}
	report := response(0x1f, 0xa4, want)
	got, err := ParseResponse(report, 0xa4)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("output = %x, want %x", got, want)
	}
}

func TestParseResponseRejectsInvalidFrames(t *testing.T) {
	valid := response(0x1f, 0x24, [6]byte{1, 2, 3, 4})
	tests := map[string]func(*Report){
		"status":   func(r *Report) { r[0] = 0x05; r[88] = checksum(*r) },
		"class":    func(r *Report) { r[6] = 0x01; r[88] = checksum(*r) },
		"echo":     func(r *Report) { r[7] = 0xa4; r[88] = checksum(*r) },
		"length":   func(r *Report) { r[5] = 5; r[88] = checksum(*r) },
		"checksum": func(r *Report) { r[88] ^= 0xff },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			report := valid
			mutate(&report)
			_, err := ParseResponse(report, 0x24)
			if err == nil {
				t.Fatal("accepted invalid response")
			}
		})
	}
}

func response(transactionID, command byte, output [6]byte) Report {
	report := BuildRequest(transactionID, command, output)
	report[0] = successStatus
	report[88] = checksum(report)
	return report
}
