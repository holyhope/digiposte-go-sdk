package chrome_test

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestRcloneDigiposteLogin(t *testing.T) {
	t.Parallel()

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "RcloneDigiposteLogin Suite")
}
