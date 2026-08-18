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

	keyboard, err := provider.Open(ctx, profile.Keyboard)
	if err != nil {
		return fmt.Errorf("open wired keyboard: %w", err)
	}
	defer keyboard.Close()

	var zero [6]byte
	receiverID, err := receiver.Command(ctx, profile.Commands.ReceiverIdentity, zero)
	if err != nil {
		return fmt.Errorf("read receiver identity (0x%02x): %w", profile.Commands.ReceiverIdentity, err)
	}
	handshake, err := keyboard.Command(ctx, profile.Commands.KeyboardPrepare, receiverID)
	if err != nil {
		return fmt.Errorf("prepare keyboard (0x%02x): %w", profile.Commands.KeyboardPrepare, err)
	}

	if confirm == nil || !confirm() {
		return ErrCancelled
	}

	commit, err := keyboard.Command(ctx, profile.Commands.KeyboardCommit, zero)
	if err != nil {
		return fmt.Errorf("commit keyboard pairing (0x%02x): %w", profile.Commands.KeyboardCommit, err)
	}
	if !bytes.Equal(handshake[:4], commit[:4]) {
		return ErrVerificationMismatch
	}
	return nil
}
