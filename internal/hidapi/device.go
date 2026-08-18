//go:build cgo && (darwin || linux || windows || freebsd)

package hidapi

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	gohid "github.com/sstallion/go-hid"

	"razer-pair/internal/protocol"
)

const (
	commandAttempts = 10
	commandDelay    = 200 * time.Millisecond
)

type device struct {
	mu            sync.Mutex
	handle        *gohid.Device
	transactionID byte
}

func (d *device) Command(ctx context.Context, command byte, input [6]byte) ([6]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.handle == nil {
		return [6]byte{}, errors.New("HID device is closed")
	}

	request := encodeFeatureReport(protocol.BuildRequest(d.transactionID, command, input))
	var lastErr error
	for attempt := 1; attempt <= commandAttempts; attempt++ {
		if err := d.send(request); err != nil {
			return [6]byte{}, err
		}
		if err := wait(ctx, commandDelay); err != nil {
			return [6]byte{}, err
		}
		response, err := d.receive()
		if err != nil {
			lastErr = err
			continue
		}
		output, err := protocol.ParseResponse(response, command)
		if err == nil {
			return output, nil
		}
		lastErr = fmt.Errorf("attempt %d: %w", attempt, err)
		if terminalStatus(response[0]) {
			break
		}
	}
	return [6]byte{}, lastErr
}

func (d *device) send(report [wireReportSize]byte) error {
	n, err := d.handle.SendFeatureReport(report[:])
	if err != nil {
		return fmt.Errorf("send feature report: %w", err)
	}
	if n != len(report) {
		return fmt.Errorf("feature request length: wrote %d, want %d", n, len(report))
	}
	return nil
}

func (d *device) receive() (protocol.Report, error) {
	var report [wireReportSize]byte
	n, err := d.handle.GetFeatureReport(report[:])
	if err != nil {
		return protocol.Report{}, fmt.Errorf("receive feature report: %w", err)
	}
	return decodeFeatureReport(report, n)
}

func (d *device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.handle == nil {
		return nil
	}
	err := d.handle.Close()
	d.handle = nil
	return err
}

func terminalStatus(status byte) bool {
	return status == 0x03 || status == 0x05
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
