package protocol

import "fmt"

const (
	ReportSize         = 90
	successStatus      = 0x02
	statusIndex        = 0
	transactionIndex   = 1
	payloadLengthIndex = 5
	commandClassIndex  = 6
	commandIndex       = 7
	payloadIndex       = 8
	checksumIndex      = 88
)

type Report [ReportSize]byte

func BuildRequest(transactionID, command byte, input [6]byte) Report {
	var report Report
	report[transactionIndex] = transactionID
	report[payloadLengthIndex] = byte(len(input))
	report[commandClassIndex] = 0x00
	report[commandIndex] = command
	copy(report[payloadIndex:], input[:])
	report[checksumIndex] = checksum(report)
	return report
}

func checksum(report Report) byte {
	var checksum byte
	for i := 2; i < checksumIndex; i++ {
		checksum ^= report[i]
	}
	return checksum
}

type responseError struct {
	reason string
	status byte
}

func (e *responseError) Error() string {
	if e.status != 0 {
		return fmt.Sprintf("%s (status 0x%02x)", e.reason, e.status)
	}
	return e.reason
}

func ParseResponse(report Report, command byte) ([6]byte, error) {
	if report[statusIndex] != successStatus {
		return [6]byte{}, &responseError{reason: "device rejected or has not completed the command", status: report[statusIndex]}
	}
	if report[commandClassIndex] != 0x00 {
		return [6]byte{}, &responseError{reason: fmt.Sprintf("unexpected command class 0x%02x", report[commandClassIndex])}
	}
	if report[commandIndex] != command {
		return [6]byte{}, &responseError{reason: fmt.Sprintf("command echo mismatch: got 0x%02x, want 0x%02x", report[commandIndex], command)}
	}
	if report[payloadLengthIndex] != 6 {
		return [6]byte{}, &responseError{reason: fmt.Sprintf("payload length: got %d, want 6", report[payloadLengthIndex])}
	}
	if got, want := report[checksumIndex], checksum(report); got != want {
		return [6]byte{}, &responseError{reason: fmt.Sprintf("checksum mismatch: got 0x%02x, want 0x%02x", got, want)}
	}
	var output [6]byte
	copy(output[:], report[payloadIndex:])
	return output, nil
}
