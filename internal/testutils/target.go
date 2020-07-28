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
	testTargetHosts  []string
	testTargetIQN    string
	testTgtHostLUNID string
	testInitiatorIQN string
)

// IntegrationTestTargetRunner is setup function for iSCSI target
func IntegrationTestTargetRunner(m *testing.M) int {
	realTargetAddress := os.Getenv("OS_BRICK_TEST_PORTAL_ADDRESS")
	realTargetIQN := os.Getenv("OS_BRICK_TEST_TARGET_IQN")

	if realTargetAddress != "" {
		// connect real target portal address
		log.Printf("test endpoint: %s", realTargetAddress)
		return integrationTestTargetRunnerReal(m, realTargetAddress, realTargetIQN)
	}

	return integrationTestTargetRunnerVirtual(m)
}

// GetTestTarget return portalIPs, targetIQN, teardown function
func GetTestTarget() ([]string, string, func()) {
	if len(testTargetHosts) == 0 {
		panic("testTarget is not initialized yes")
	}

	return testTargetHosts, testTargetIQN, func() {}
}
