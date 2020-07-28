package testutils

import (
	"os"
	"testing"
)

// testing file path
const (
	testDockerfilePath = "%s/../../test/docker/Dockerfile"
)

var (
	testTargetHosts   []string
	testTargetIQN     string
	testTgtHostLUNID  string
	testInitiatorIQN  string
	realTargetAddress string
	realTargetIQN     string
)

func init() {
	realTargetAddress = os.Getenv("OS_BRICK_TEST_PORTAL_ADDRESS")
	realTargetIQN = os.Getenv("OS_BRICK_TEST_TARGET_IQN")
}

// IntegrationTestTargetRunner is setup function for iSCSI target
func IntegrationTestTargetRunner(m *testing.M) int {
	if realTargetAddress != "" {
		// connect real target portal address
		return integrationTestTargetRunnerReal(m)
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
