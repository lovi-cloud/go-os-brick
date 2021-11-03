package osbrick

import (
	"context"
	"testing"

	"github.com/lovi-cloud/go-os-brick/internal/testutils"
)

func TestConnectMultiPathVolume(t *testing.T) {
	testTargetIPs, testTargetIQN, teardown := testutils.GetTestTarget()
	defer teardown()

	if len(testTargetIPs) == 1 {
		t.Skip("not implement multipath")
	}

	testTargetIQNs := make([]string, len(testTargetIPs))
	for i := 0; i<len(testTargetIPs); i++ {
		testTargetIQNs = append(testTargetIQNs, testTargetIQN)
	}

	// NOTE(whywaita): testing volume hostlunid is 1 to 10
	for i := 1; i <= 10; i++ {
		hostLUNID := i

		deviceName, err := ConnectMultiPathVolume(context.Background(), testTargetIPs, testTargetIQNs, hostLUNID)
		if err != nil {
			t.Errorf("ConnectMultipathVolume return err: %+v", err)
		}

		t.Logf("found device name: %s\n", deviceName)
	}
}

func TestDisconnectVolume(t *testing.T) {
	testTargetIPs, testTargetIQN, teardown := testutils.GetTestTarget()
	defer teardown()

	if len(testTargetIPs) == 1 {
		// not implement multipath
		t.Skip("not implement multipath")
	}

	testTargetIQNs := make([]string, len(testTargetIPs))
	for i := 0; i<len(testTargetIPs); i++ {
		testTargetIQNs = append(testTargetIQNs, testTargetIQN)
	}

	// NOTE(whywaita): testing volume hostlunid is 1 to 10
	for i := 1; i <= 10; i++ {
		hostLUNID := i

		if err := DisconnectVolume(context.Background(), testTargetIPs,testTargetIQNs, hostLUNID); err != nil {
			t.Errorf("DisconnectVolume return err: %+v", err)
		}
	}
}
