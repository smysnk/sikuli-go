//go:build darwin

package grpcv1

import (
	"os"
	"testing"
)

func TestScreencaptureArgsDefaults(t *testing.T) {
	t.Setenv("SIKULI_CAPTURE_DISPLAY", "")
	t.Setenv("SIKULIGO_CAPTURE_DISPLAY", "")
	args := screencaptureArgs("/tmp/capture.png")
	want := []string{"-x", "-t", "png", "/tmp/capture.png"}
	if len(args) != len(want) {
		t.Fatalf("arg length mismatch: got=%v want=%v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("arg mismatch at %d: got=%q want=%q (all=%v)", i, args[i], want[i], args)
		}
	}
}

func TestScreencaptureArgsUsesPrimaryDisplayEnv(t *testing.T) {
	t.Setenv("SIKULI_CAPTURE_DISPLAY", "2")
	t.Setenv("SIKULIGO_CAPTURE_DISPLAY", "")
	args := screencaptureArgs("/tmp/capture.png")
	want := []string{"-x", "-t", "png", "-D", "2", "/tmp/capture.png"}
	if len(args) != len(want) {
		t.Fatalf("arg length mismatch: got=%v want=%v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("arg mismatch at %d: got=%q want=%q (all=%v)", i, args[i], want[i], args)
		}
	}
}

func TestScreencaptureArgsUsesFallbackDisplayEnv(t *testing.T) {
	t.Setenv("SIKULI_CAPTURE_DISPLAY", "")
	t.Setenv("SIKULIGO_CAPTURE_DISPLAY", "5")
	args := screencaptureArgs("/tmp/capture.png")
	found := false
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "-D" && args[i+1] == "5" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected -D 5 in args, got=%v", args)
	}
}

func TestScreencaptureArgsPrimaryEnvWins(t *testing.T) {
	t.Setenv("SIKULI_CAPTURE_DISPLAY", "3")
	t.Setenv("SIKULIGO_CAPTURE_DISPLAY", "7")
	args := screencaptureArgs("/tmp/capture.png")
	display := ""
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "-D" {
			display = args[i+1]
			break
		}
	}
	if display != "3" {
		t.Fatalf("expected display selector 3, got=%q args=%v", display, args)
	}
}

func TestScreencaptureArgsTrimsWhitespace(t *testing.T) {
	t.Setenv("SIKULI_CAPTURE_DISPLAY", "  11  ")
	t.Setenv("SIKULIGO_CAPTURE_DISPLAY", "")
	args := screencaptureArgs("/tmp/capture.png")
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "-D" {
			if args[i+1] != "11" {
				t.Fatalf("expected trimmed display selector 11, got=%q", args[i+1])
			}
			return
		}
	}
	t.Fatalf("expected -D arg, got=%v", args)
}

func TestScreencaptureArgsNoProcessEnvLeakInTest(t *testing.T) {
	_ = os.Getenv("SIKULI_CAPTURE_DISPLAY")
	_ = os.Getenv("SIKULIGO_CAPTURE_DISPLAY")
}
