package osbrick

import (
	"context"
	"testing"

	"github.com/whywaita/go-os-brick/internal/testutils"
)

func TestLoginLogoutPortal(t *testing.T) {
	testTargetIPs, targetIQN, teardown := testutils.GetTestTarget()
	defer teardown()

	for _, testTargetIP := range testTargetIPs {
		_, err := doSendtargets(context.Background(), testTargetIP)
		if err != nil {
			t.Errorf("doSendtargets return err: %+v", err)
		}

		if err := LoginPortal(context.Background(), testTargetIP, targetIQN); err != nil {
			t.Errorf("LoginPortal return err: %+v", err)
		}

		if err := LogoutPortal(context.Background(), testTargetIP, targetIQN); err != nil {
			t.Errorf("LogoutPortal return err: %+v", err)
		}
	}
}

func TestDoSendtargets(t *testing.T) {
	testTargetIPs, _, teardown := testutils.GetTestTarget()
	defer teardown()

	for _, testTargetIP := range testTargetIPs {
		_, err := doSendtargets(context.Background(), testTargetIP)
		if err != nil {
			t.Errorf("doSendTarget return err: %+v", err)
		}
	}
}
