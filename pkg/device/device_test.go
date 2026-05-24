package device

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHardwareIDToUUIDIsDeterministicAndNormalized(t *testing.T) {
	first := hardwareIDToUUID("mainboard", " abc   123 ")
	second := hardwareIDToUUID(" MAINBOARD ", "ABC 123")

	if first != second {
		t.Fatalf("expected normalized hardware values to generate the same UUID, got %q and %q", first, second)
	}
	if !isValidUUID(first) {
		t.Fatalf("expected a valid UUID, got %q", first)
	}
}

func TestIsUsableHardwareIDRejectsPlaceholders(t *testing.T) {
	invalidValues := []string{
		"",
		"0",
		"0000-0000-0000",
		"FFFFFFFF-FFFF-FFFF",
		"To Be Filled By O.E.M.",
		"Default string",
		"Not Specified",
		"System Serial Number",
		"N/A",
	}

	for _, value := range invalidValues {
		if isUsableHardwareID(value) {
			t.Fatalf("expected %q to be rejected", value)
		}
	}

	if !isUsableHardwareID("MB-123456789") {
		t.Fatal("expected a concrete mainboard serial to be accepted")
	}
}

func TestFirstUsableOutputLineSkipsHeadersAndInvalidValues(t *testing.T) {
	output := []byte("SerialNumber\r\nTo Be Filled By O.E.M.\r\n MB-123456 \r\n")

	got := firstUsableOutputLine(output, "SerialNumber")
	if got != "MB-123456" {
		t.Fatalf("expected first usable serial, got %q", got)
	}
}

func TestReadFirstUsableFileSkipsInvalidFiles(t *testing.T) {
	dir := t.TempDir()
	invalidPath := filepath.Join(dir, "invalid")
	validPath := filepath.Join(dir, "valid")

	if err := os.WriteFile(invalidPath, []byte("Not Specified\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(validPath, []byte("mb-987654\n"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := readFirstUsableFile(invalidPath, validPath)
	if err != nil {
		t.Fatalf("expected usable serial from second file: %v", err)
	}
	if got != "MB-987654" {
		t.Fatalf("expected canonical serial, got %q", got)
	}
}
