package chrome

import "testing"

func TestWithScreenDumpDir(t *testing.T) {
	t.Parallel()

	t.Run("sets screenDumpDir on chromeLogin", func(t *testing.T) {
		t.Parallel()

		instance := &chromeLogin{
			url:                "",
			cookies:            nil,
			screenShortOnError: false,
			refreshFrequency:   0,
			timeout:            0,
			screenDumpDir:      "",
			infoLogger:         nil,
			errorLogger:        nil,
			binaryPath:         "",
		}

		err := WithScreenDumpDir("/tmp/dump").Apply(instance)
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}

		if instance.screenDumpDir != "/tmp/dump" {
			t.Fatalf("screenDumpDir = %q, want %q", instance.screenDumpDir, "/tmp/dump")
		}
	})

	t.Run("rejects empty dir like other options reject invalid input", func(t *testing.T) {
		t.Parallel()

		validatable, ok := WithScreenDumpDir("").(Validatable)
		if !ok {
			t.Fatal("WithScreenDumpDir(\"\") does not implement Validatable")
		}

		err := validatable.Validate()
		if err == nil {
			t.Fatal("expected an error for an empty dir")
		}
	})

	t.Run("rejects wrong instance type", func(t *testing.T) {
		t.Parallel()

		err := WithScreenDumpDir("/tmp/dump").Apply(&struct{}{})
		if err == nil {
			t.Fatal("expected an error for a non-*chromeLogin instance")
		}
	})
}
