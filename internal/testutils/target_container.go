// +build container

package testutils

import (
	"errors"
	"fmt"
	"log"
	"net"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func integrationTestTargetRunnerVirtual(m *testing.M) int {
	testTargetIQN = "iqn.0000-00.com.example:target0"
	testTgtHostLUNID = "0"
	testInitiatorIQN = "iqn.0000-00.com.example:initiator0"

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %+v", err)
	}

	options := &dockertest.RunOptions{
		Name:       "iscsi-target",
		Privileged: true,
		PortBindings: map[docker.Port][]docker.PortBinding{
			"3260/tcp": {{HostPort: "3260/tcp"}},
		},
	}

	_, pwd, _, _ := runtime.Caller(0)
	resource, err := pool.BuildAndRunWithOptions(fmt.Sprintf(testDockerfilePath, path.Dir(pwd)), options)
	if err != nil {
		log.Fatalf("Could not start resource: %+v", err)
	}

	if err := pool.Retry(func() error {
		targetHost := fmt.Sprintf("%s:%s", resource.Container.NetworkSettings.IPAddress, "3260")

		conn, err := net.DialTimeout("tcp", targetHost, 1*time.Second)
		if err != nil {
			return err
		}
		if conn == nil {
			return errors.New("could not create connection to docker")
		}
		defer conn.Close()

		testTargetHost = targetHost
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %+v", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %+v", err)
	}

	return code
}

// GetTestTarget return portalIP, targetIQN, teardown function
func GetTestTarget() (string, string func()) {
	if testTargetHost == "" {
		panic("testTarget is not initialized yes")
	}

	return testTargetHost, testTargetIQN, func() { truncateDisk() }
}

func truncateDisk() {

}
