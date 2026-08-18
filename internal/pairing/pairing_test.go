package pairing_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"razer-pair/internal/hid"
	"razer-pair/internal/mockhid"
	"razer-pair/internal/model"
	"razer-pair/internal/pairing"
)

func TestPairSuccessConsumesVerifiedTranscript(t *testing.T) {
	provider := scenario(t, "success")
	profile := model.Default()
	confirmed := false

	err := pairing.Pair(context.Background(), provider, profile, func() bool {
		confirmed = true
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if !confirmed {
		t.Fatal("confirmation callback was not called")
	}
	for _, role := range []hid.Role{hid.Receiver, profile.Peripheral.Role} {
		if remaining := provider.Remaining(role); remaining != 0 {
			t.Fatalf("%s has %d unconsumed exchanges", role, remaining)
		}
	}
}

func TestPairCancellationDoesNotCommit(t *testing.T) {
	provider := scenario(t, "success")
	err := pairing.Pair(context.Background(), provider, model.Default(), func() bool { return false })
	if !errors.Is(err, pairing.ErrCancelled) {
		t.Fatalf("error = %v, want ErrCancelled", err)
	}
	if remaining := provider.Remaining(model.Default().Peripheral.Role); remaining != 1 {
		t.Fatalf("device remaining exchanges = %d, want 1 commit", remaining)
	}
}

func TestPairRejectsPostCommitMismatch(t *testing.T) {
	provider := scenario(t, "mismatch")
	err := pairing.Pair(context.Background(), provider, model.Default(), func() bool { return true })
	if !errors.Is(err, pairing.ErrVerificationMismatch) {
		t.Fatalf("error = %v, want ErrVerificationMismatch", err)
	}
}

func TestPairReportsMissingDevice(t *testing.T) {
	provider := scenario(t, "missing-device")
	err := pairing.Pair(context.Background(), provider, model.Default(), func() bool { return true })
	if err == nil || !strings.Contains(err.Error(), "wired keyboard") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func scenario(t *testing.T, name string) *mockhid.Provider {
	t.Helper()
	provider, err := mockhid.Scenario(name, model.Default())
	if err != nil {
		t.Fatal(err)
	}
	return provider
}
