package osbrick

import (
	"os"
	"testing"

	"github.com/lovi-cloud/go-os-brick/internal/testutils"
)

func TestMain(m *testing.M) {
	os.Exit(testutils.IntegrationTestTargetRunner(m))
}
