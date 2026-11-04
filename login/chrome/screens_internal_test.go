package chrome

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// withTestLoggers attaches discard-output loggers, since infoLogger and
// errorLogger panic if none is set on ctx.
func withTestLoggers(ctx context.Context) context.Context {
	discard := log.New(io.Discard, "", 0)

	return withErrorLogger(withInfoLogger(ctx, discard), discard)
}

// fakeScreen is a minimal Screen that never touches CDP, so it can be run
// through Screens.run without a live browser target as long as resolution
// itself is driven by a fake screenResolver instead of the real resolve().
type fakeScreen struct {
	matches atomic.Bool
}

func (f *fakeScreen) String() string { return "fake" }

func (f *fakeScreen) CurrentPageMatches(context.Context) bool { return f.matches.Load() }

func (f *fakeScreen) ShouldWaitForResponse() bool { return false }

func (f *fakeScreen) Do(context.Context) error { return nil }

// fakeResolver is a screenResolver that never touches CDP; it just counts
// calls and reports success, so Screens.run can be exercised end to end in
// a unit test without a live browser target.
type fakeResolver struct {
	calls atomic.Int32
}

func (f *fakeResolver) resolve(context.Context, Screen) error {
	f.calls.Add(1)

	return nil
}

// spyCapturer is a networkCapturer that records whether/how many times it
// was called, without touching CDP.
type spyCapturer struct {
	calls atomic.Int32
	dirs  []string
}

func (s *spyCapturer) capture(_ context.Context, dir string) error {
	s.calls.Add(1)
	s.dirs = append(s.dirs, dir)

	return nil
}

func TestScreensRunUsesInjectedResolver(t *testing.T) {
	t.Parallel()

	screen := &fakeScreen{matches: atomic.Bool{}}
	resolver := &fakeResolver{calls: atomic.Int32{}}

	screens := &Screens{
		screens:          []Screen{screen},
		refreshFrequency: time.Millisecond,
		resolver:         resolver,
		succeeded:        atomic.Bool{},
	}

	ctx, cancel := context.WithTimeout(withTestLoggers(t.Context()), 200*time.Millisecond)
	defer cancel()

	screen.matches.Store(true)

	go screens.run(ctx, screen)

	waitFor(t, func() bool { return resolver.calls.Load() > 0 })

	screens.succeeded.Store(true)
}

func TestChromeLoginMaybeStartCaptureWhenScreenDumpDirSet(t *testing.T) {
	t.Parallel()

	capturer := &spyCapturer{calls: atomic.Int32{}, dirs: nil}
	instance := &chromeLogin{
		url:                "",
		cookies:            nil,
		screenShortOnError: false,
		refreshFrequency:   0,
		timeout:            0,
		screenDumpDir:      t.TempDir(),
		capturer:           capturer,
		infoLogger:         nil,
		errorLogger:        nil,
		binaryPath:         "",
	}

	instance.maybeStartCapture(withTestLoggers(t.Context()))

	if capturer.calls.Load() != 1 {
		t.Fatalf("capturer called %d times, want 1", capturer.calls.Load())
	}

	if got := capturer.dirs[0]; got != instance.screenDumpDir {
		t.Fatalf("capture called with dir %q, want %q", got, instance.screenDumpDir)
	}
}

func TestChromeLoginMaybeStartCaptureDoesNotCaptureWhenScreenDumpDirUnset(t *testing.T) {
	t.Parallel()

	capturer := &spyCapturer{calls: atomic.Int32{}, dirs: nil}
	instance := &chromeLogin{
		url:                "",
		cookies:            nil,
		screenShortOnError: false,
		refreshFrequency:   0,
		timeout:            0,
		screenDumpDir:      "",
		capturer:           capturer,
		infoLogger:         nil,
		errorLogger:        nil,
		binaryPath:         "",
	}

	instance.maybeStartCapture(withTestLoggers(t.Context()))

	if capturer.calls.Load() != 0 {
		t.Fatalf("expected capturer to never be called, got %d calls", capturer.calls.Load())
	}
}

func TestWriteCapturedResponse(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	tests := []struct {
		name     string
		rawURL   string
		wantPath string
	}{
		{
			name:     "root path",
			rawURL:   "https://example.com/",
			wantPath: filepath.Join("example.com", "index.html"),
		},
		{
			name:     "empty path",
			rawURL:   "https://example.com",
			wantPath: filepath.Join("example.com", "index.html"),
		},
		{
			name:     "regular path",
			rawURL:   "https://example.com/login",
			wantPath: filepath.Join("example.com", "login"),
		},
		{
			name:     "nested asset path",
			rawURL:   "https://assets.example.com/static/app.css",
			wantPath: filepath.Join("assets.example.com", "static", "app.css"),
		},
		{
			name:     "traversal attempt is contained",
			rawURL:   "https://example.com/../../etc/passwd",
			wantPath: filepath.Join("example.com", "etc", "passwd"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body := []byte("body for " + testCase.name)

			subdir := filepath.Join(dir, testCase.name)

			err := writeCapturedResponse(subdir, testCase.rawURL, body)
			if err != nil {
				t.Fatalf("writeCapturedResponse: %v", err)
			}

			got, err := os.ReadFile(filepath.Join(subdir, testCase.wantPath))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			if string(got) != string(body) {
				t.Fatalf("content = %q, want %q", got, body)
			}
		})
	}
}

func TestWriteCapturedResponseCollisionAppendsSuffix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	err := writeCapturedResponse(dir, "https://example.com/login-actions/authenticate", []byte("credentials"))
	if err != nil {
		t.Fatalf("writeCapturedResponse (1st): %v", err)
	}

	err = writeCapturedResponse(dir, "https://example.com/login-actions/authenticate", []byte("otp"))
	if err != nil {
		t.Fatalf("writeCapturedResponse (2nd): %v", err)
	}

	err = writeCapturedResponse(dir, "https://example.com/login-actions/authenticate", []byte("trusted-device"))
	if err != nil {
		t.Fatalf("writeCapturedResponse (3rd): %v", err)
	}

	for path, want := range map[string]string{
		"login-actions/authenticate":   "credentials",
		"login-actions/authenticate-2": "otp",
		"login-actions/authenticate-3": "trusted-device",
	} {
		got, err := os.ReadFile(filepath.Join(dir, "example.com", path))
		if err != nil {
			t.Fatalf("ReadFile(%q): %v", path, err)
		}

		if string(got) != want {
			t.Fatalf("content of %q = %q, want %q", path, got, want)
		}
	}
}

func TestWriteCapturedResponseRejectsHostlessURL(t *testing.T) {
	t.Parallel()

	err := writeCapturedResponse(t.TempDir(), "not-a-url-with-no-host", []byte("x"))
	if err == nil {
		t.Fatal("expected an error for a URL with no hostname")
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if cond() {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("condition was never met")
}
