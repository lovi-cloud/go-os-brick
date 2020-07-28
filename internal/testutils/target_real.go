package testutils

import (
	"strings"
	"testing"
)

func integrationTestTargetRunnerReal(m *testing.M, realTargetAddress, realTargetIQN string) int {
	addresses := strings.Split(realTargetAddress, ",")

	testTargetHosts = addresses
	testTargetIQN = realTargetIQN

	code := m.Run()

	return code
}
