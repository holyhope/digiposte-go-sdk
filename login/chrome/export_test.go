package chrome

// This file exists only to expose test-only, maintainer-facing capture
// tooling to this package's own external test package (chrome_test). Being
// a _test.go file, it is excluded from `go build`/`go doc` of the module:
// WithScreenDumpDir never ships as part of the public API. See
// openspec/changes/mock-chrome-login-screens/design.md, Decision 1.

import (
	"errors"

	login "github.com/holyhope/digiposte-go-sdk/login"
)

var errEmptyScreenDumpDir = errors.New("dir is empty")

// WithScreenDumpDir is a maintainer-only fixture-capture helper: when set,
// every CDP network response (documents and the assets they load) observed
// while resolving a screen is recorded under dir, organized by the real
// hostname and URL path that served it. It exists to (re)generate this
// package's offline test fixtures from a real login session; it has no use
// to, and is not reachable by, a library consumer.
func WithScreenDumpDir(dir string) login.Option { //nolint:ireturn
	return &withScreenDumpDir{Dir: dir}
}

type withScreenDumpDir struct {
	Dir string
}

func (o *withScreenDumpDir) Apply(instance any) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.screenDumpDir = o.Dir

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

func (o *withScreenDumpDir) Validate() error {
	if o.Dir == "" {
		return &login.InvalidOptionError{
			Name: "WithScreenDumpDir",
			Err:  errEmptyScreenDumpDir,
		}
	}

	return nil
}
