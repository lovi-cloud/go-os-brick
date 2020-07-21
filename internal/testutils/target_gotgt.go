// +build !container

package testutils

import "testing"

func integrationTestTargetRunnerVirtual(m *testing.M) int {
	// TODO: add test
	return 0
}

// GetTestTargetAddress return target address for testing
func GetTestTargetAddress() (string, func()) {
	return "127.0.0.1:3260", func() {}
}
