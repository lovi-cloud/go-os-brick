package osbrick

import (
	"context"
	"testing"

	"github.com/lovi-cloud/go-os-brick/internal/testutils"
)

func TestConnectMultiPathVolume(t *testing.T) {
	testTargetIPs, _, teardown := testutils.GetTestTarget()
	defer teardown()

	if len(testTargetIPs) == 1 {
		t.Skip("not implement multipath")
	}

	// NOTE(whywaita): testing volume hostlunid is 1 to 10
	for i := 1; i <= 10; i++ {
		hostLUNID := i

		deviceName, err := ConnectMultiPathVolume(context.Background(), testTargetIPs, hostLUNID)
		if err != nil {
			t.Errorf("ConnectMultipathVolume return err: %+v", err)
		}

		t.Logf("found device name: %s\n", deviceName)
	}
}

func TestDisconnectVolume(t *testing.T) {
	testTargetIPs, _, teardown := testutils.GetTestTarget()
	defer teardown()

	if len(testTargetIPs) == 1 {
		// not implement multipath
		t.Skip("not implement multipath")
	}

	// NOTE(whywaita): testing volume hostlunid is 1 to 10
	for i := 1; i <= 10; i++ {
		hostLUNID := i

		if err := DisconnectVolume(context.Background(), testTargetIPs, hostLUNID); err != nil {
			t.Errorf("DisconnectVolume return err: %+v", err)
		}
	}
}
