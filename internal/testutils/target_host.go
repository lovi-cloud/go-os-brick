// +build host

package testutils

import (
	"context"
	"log"
	"os/exec"
	"testing"
)

func integrationTestTargetRunnerVirtual(m *testing.M) int {
	testTargetIQN = "iqn.0000-00.com.example:target0"
	testTgtHostLUNID = "0"
	testInitiatorIQN = "iqn.0000-00.com.example:initiator0"
	testTargetHosts = []string{"127.0.0.1:3260"}

	if out, err := exec.CommandContext(context.Background(), "../test/scripts/init.sh").CombinedOutput(); err != nil {
		log.Printf("init.sh return err: %+v (out: %+v)", err, out)
		return 1
	}

	code := m.Run()

	if out, err := exec.CommandContext(context.Background(), "../test/scripts/teardown.sh").CombinedOutput(); err != nil {
		log.Printf("init.sh return err: %+v (out: %+v)", err, out)
		return 1
	}

	return code
}