package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"razer-pair/internal/cli"
)

func run(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := cli.Run(context.Background(), args, cli.Options{
		Version: "test", Stdin: strings.NewReader(stdin), Stdout: &stdout, Stderr: &stderr,
	})
	return code, stdout.String(), stderr.String()
}

func TestInspectMock(t *testing.T) {
	code, stdout, stderr := run(t, "", "--mock", "success", "inspect")
	if code != cli.ExitOK {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
	if !strings.Contains(stdout, "1532:0277 verified") || !strings.Contains(stdout, "1532:027b verified") || !strings.Contains(stdout, "Ready:") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}
}

func TestInspectVerboseShowsCollectionDetails(t *testing.T) {
	code, stdout, stderr := run(t, "", "--mock", "success", "inspect", "--verbose")
	if code != cli.ExitOK {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
	if !strings.Contains(stdout, "HID collections") || !strings.Contains(stdout, "usage=0x0001:0x0002") || !strings.Contains(stdout, "mock non-pairing collection") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}
}

func TestScanMock(t *testing.T) {
	code, stdout, stderr := run(t, "", "--model", "basilisk-ultimate", "--mock", "success", "scan")
	if code != cli.ExitOK {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
	if !strings.Contains(stdout, "1532:0086 feature=90") || !strings.Contains(stdout, "1532:0088 feature=90") || !strings.Contains(stdout, "No reports sent") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}
}

func TestDryRunSendsNothingAndPrintsSequence(t *testing.T) {
	code, stdout, stderr := run(t, "", "--mock", "success", "dry-run")
	if code != cli.ExitOK {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
	if !strings.Contains(stdout, "No reports sent") || !strings.Contains(stdout, "0x95") || !strings.Contains(stdout, "0xa4") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}
}

func TestPairYes(t *testing.T) {
	code, stdout, stderr := run(t, "", "--mock", "success", "pair", "--yes")
	if code != cli.ExitOK {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
	if !strings.Contains(stdout, "Pairing successful") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}
}

func TestPairInteractiveCancellation(t *testing.T) {
	code, _, stderr := run(t, "n\n", "--mock", "success", "pair")
	if code != cli.ExitCancelled || !strings.Contains(stderr, "commit command was not sent") {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
}

func TestPairInteractiveConfirmation(t *testing.T) {
	code, stdout, stderr := run(t, "yes\n", "--mock", "success", "pair")
	if code != cli.ExitOK || !strings.Contains(stdout, "Pairing successful") {
		t.Fatalf("code = %d, stdout = %s, stderr = %s", code, stdout, stderr)
	}
}

func TestPairMismatchFails(t *testing.T) {
	code, _, stderr := run(t, "", "--mock", "mismatch", "pair", "--yes")
	if code != cli.ExitDevice || !strings.Contains(stderr, "identity did not match") {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
}

func TestMissingDeviceFailsInspect(t *testing.T) {
	code, _, stderr := run(t, "", "--mock", "missing-device", "inspect")
	if code != cli.ExitDevice || !strings.Contains(stderr, "wired keyboard not detected") || !strings.Contains(stderr, "--verbose") {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
}

func TestAccessDeniedFailsBeforePairing(t *testing.T) {
	code, _, stderr := run(t, "", "--mock", "denied", "pair", "--yes")
	if code != cli.ExitDevice || !strings.Contains(stderr, "access denied") {
		t.Fatalf("code = %d, stderr = %s", code, stderr)
	}
}

func TestListModels(t *testing.T) {
	code, stdout, stderr := run(t, "", "list-models")
	if code != cli.ExitOK || !strings.Contains(stdout, "pro-type-ultra") || !strings.Contains(stdout, "basilisk-ultimate") {
		t.Fatalf("code = %d, stdout = %s, stderr = %s", code, stdout, stderr)
	}
}

func TestBasiliskMockPair(t *testing.T) {
	code, stdout, stderr := run(t, "", "--model", "basilisk-ultimate", "--mock", "success", "pair", "--yes")
	if code != cli.ExitOK || !strings.Contains(stdout, "Pairing successful") {
		t.Fatalf("code = %d, stdout = %s, stderr = %s", code, stdout, stderr)
	}
}
