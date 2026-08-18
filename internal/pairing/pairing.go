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
	ErrVerificationMismatch = errors.New("post-commit receiver identity did not match")
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
	receiverID, err := receiver.Command(ctx, profile.Commands.ReceiverIdentity, zero)
	if err != nil {
		return fmt.Errorf("read receiver identity (0x%02x): %w", profile.Commands.ReceiverIdentity, err)
	}
	handshake, err := peripheral.Command(ctx, profile.Commands.PeripheralPrepare, receiverID)
	if err != nil {
		return fmt.Errorf("prepare %s (0x%02x): %w", profile.Peripheral.Label, profile.Commands.PeripheralPrepare, err)
	}

	if confirm == nil || !confirm() {
		return ErrCancelled
	}

	commit, err := peripheral.Command(ctx, profile.Commands.PeripheralCommit, zero)
	if err != nil {
		return fmt.Errorf("commit %s pairing (0x%02x): %w", profile.Peripheral.Label, profile.Commands.PeripheralCommit, err)
	}
	if !bytes.Equal(handshake[:4], commit[:4]) {
		return ErrVerificationMismatch
	}
	return nil
}
