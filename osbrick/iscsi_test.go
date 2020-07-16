package osbrick

import (
	"context"
	"testing"

	"github.com/whywaita/go-os-brick/internal/testutils"
)

func TestLoginPortal(t *testing.T) {
	portalIP, targetIQN, teardown, err := testutils.GetTestTarget()
	if err != nil {
		t.Fatalf("failed to get test target %+v\n", err)
	}
	defer teardown()

	_, err = doSendtargets(context.Background(), portalIP)
	if err != nil {
		t.Fatalf("failed to sendtargets: %+v", err)
	}

	err = LoginPortal(context.Background(), portalIP, targetIQN)
	if err != nil {
		t.Fatalf("%+v\n", err)
	}
}
