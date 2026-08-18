package pairing

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"razer-pair/internal/hid"
	"razer-pair/internal/model"
)

var (
	ErrCancelled            = errors.New("pairing cancelled before commit")
	ErrVerificationMismatch = errors.New("post-commit pairing identity did not match")
)

type Confirm func() bool

func Pair(ctx context.Context, provider hid.Provider, profile model.Profile, confirm Confirm) error {
	receiver, err := provider.Open(ctx, profile.Receiver)
	if err != nil {
		return fmt.Errorf("open receiver: %w", err)
	}
	defer receiver.Close()

	peripheral, err := provider.Open(ctx, profile.Peripheral)
	if err != nil {
		return fmt.Errorf("open %s: %w", profile.Peripheral.Label, err)
	}
	defer peripheral.Close()

	var zero [6]byte
	devices := map[hid.Role]hid.Device{
		hid.Peripheral: peripheral,
		hid.Receiver:   receiver,
	}
	identity, err := run(ctx, devices, profile.Commands.Identity, zero)
	if err != nil {
		return commandError("read pairing identity", profile, profile.Commands.Identity, err)
	}
	handshake, err := run(ctx, devices, profile.Commands.Prepare, identity)
	if err != nil {
		return commandError("prepare pairing", profile, profile.Commands.Prepare, err)
	}

	if confirm == nil || !confirm() {
		return ErrCancelled
	}

	commit, err := run(ctx, devices, profile.Commands.Commit, zero)
	if err != nil {
		return commandError("commit pairing", profile, profile.Commands.Commit, err)
	}
	if !bytes.Equal(handshake[:4], commit[:4]) {
		return ErrVerificationMismatch
	}
	return nil
}

func run(ctx context.Context, devices map[hid.Role]hid.Device, command model.Command, input [6]byte) ([6]byte, error) {
	device := devices[command.Target]
	if device == nil {
		return [6]byte{}, fmt.Errorf("invalid command target %q", command.Target)
	}
	return device.Command(ctx, command.ID, input)
}

func commandError(action string, profile model.Profile, command model.Command, err error) error {
	spec, ok := profile.Spec(command.Target)
	if !ok {
		return fmt.Errorf("%s: invalid command target %q", action, command.Target)
	}
	return fmt.Errorf("%s on %s (0x%02x): %w", action, spec.Label, command.ID, err)
}
