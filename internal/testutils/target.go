package testutils

import (
	"log"
	"os"
	"testing"
)

// testing file path
const (
	testDockerfilePath = "%s/../../test/docker/Dockerfile"
)

var (
	testTargetHost   string
	testTargetIQN    string
	testTgtHostLUNID string
	testInitiatorIQN string
)

// IntegrationTestTargetRunner is setup function for iSCSI target
func IntegrationTestTargetRunner(m *testing.M) int {
	if os.Getenv("OS_BRICK_TEST_PORTAL_ADDRESS") != "" {
		// connect real target portal address
		return integrationTestTargetRunnerReal(m)
	}

	return integrationTestTargetRunnerVirtual(m)
}

func integrationTestTargetRunnerReal(m *testing.M) int {
	log.Fatalf("implement me")
	return 1
}
