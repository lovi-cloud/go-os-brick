package testutils

import (
	"log"
	"os"
	"testing"
)

// const for testing
const (
	testDockerfilePath = "%s/../../test/docker/Dockerfile"
	TargetIQN          = "iqn.0000-00.com.example:target0"
	TgtHostLUNID       = "0"
	InitiatorIQN       = "iqn.0000-00.com.example:initiator0"
)

var (
	testTargetHost string
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
