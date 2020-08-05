package osbrick

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Error codes returns by failure to exec iscsiadm command.
var (
	ErrSessionNotFound = errors.New("iSCSI session not found")
)

// iscsiadm exitcode meaning
const (
	ExitCodeAlreadyLogin = 15
	ExitCodeNoRecord     = 21
)

// command binary
var (
	BinaryTee = "tee"
)

// LoginPortal login to iSCSI portal
func LoginPortal(ctx context.Context, portalIP, targetIQN string) error {
	logf("start login to portal [Portal: %s]\n", portalIP)
	_, exitCode, err := iscsiadm(ctx, portalIP, targetIQN, []string{"--login"})
	if err != nil && exitCode != ExitCodeAlreadyLogin {
		return fmt.Errorf("failed to execute command that login to iscsi portal (PortalIP: %s): %w", portalIP, err)
	}

	_, _, err = iscsiadmUpdate(ctx, portalIP, targetIQN, "node.startup", "automatic", nil)
	if err != nil {
		return fmt.Errorf("failed to update node.startup to automatic: %w", err)
	}

	logf("successfully login! [Portal: %s]\n", portalIP)
	return nil

}

// LogoutPortal logout from iSCSI portal
func LogoutPortal(ctx context.Context, portalIP, targetIQN string) error {
	logf("start logout to portal [Portal: %s]\n", portalIP)
	_, _, err := iscsiadmUpdate(ctx, portalIP, targetIQN, "node.startup", "manual", nil)
	if err != nil {
		return fmt.Errorf("failed to update node.startup to manual: %w", err)
	}

	_, _, err = iscsiadm(ctx, portalIP, targetIQN, []string{"--logout"})
	if err != nil {
		return fmt.Errorf("failed to logout iscsi portal (PortalIP: %s): %w", portalIP, err)
	}

	logf("successfully logout! [Portal: %s]\n", portalIP)
	return nil
}

// GetIPsIQNsLUNs get a some information
func GetIPsIQNsLUNs(ctx context.Context, portalIP string, targetHostLUNID int) ([]string, []string, []int, error) {
	out, err := doSendtargets(ctx, portalIP)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to update discoverydb: %w", err)
	}

	ips, iqns, err := getIPsIQNs(out)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse sendtargets output: %w", err)
	}

	luns := getLUNs(targetHostLUNID, len(iqns))

	return ips, iqns, luns, nil
}

// getIpsIqns parse output of `iscsiadm -m discovery -t sendtargets`
// ex: 192.0.2.10:3260,1 iqn.0000-00.com.example:name1:192.0.2.10
func getIPsIQNs(out []byte) (ips []string, iqns []string, err error) {
	reader := bytes.NewReader(out)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		sentence := scanner.Text()
		d := strings.Split(sentence, " ")

		if len(d) == 2 && strings.HasPrefix(d[1], "iqn") {
			// ok
			ip := strings.Split(d[0], ",")
			ips = append(ips, ip[0])
			iqns = append(iqns, d[1])
		}
	}
	return ips, iqns, nil
}

func getLUNs(targetHostLUNID, targetIQNCount int) []int {
	var targetLUNs []int
	for i := 0; i < targetIQNCount; i++ {
		targetLUNs = append(targetLUNs, targetHostLUNID)
	}

	return targetLUNs
}

func doSendtargets(ctx context.Context, portalIP string) ([]byte, error) {
	logf("do sendtarget [Portal: %s]\n", portalIP)
	a := []string{"-m", "discovery", "-t", "sendtargets", "-p", portalIP}
	out, exitCode, err := iscsiadmBase(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("failed to execute discovery command (exit code: %d): %w", exitCode, err)
	}

	return out, nil
}

func scanISCSI(ctx context.Context, hctl *Hctl) error {
	path := fmt.Sprintf("/sys/class/scsi_host/host%d/scan", hctl.HostID)
	content := fmt.Sprintf("%d %d %d",
		hctl.ChannelID,
		hctl.TargetID,
		hctl.HostLUNID)

	return echoScsiCommand(ctx, path, content)
}

func echoScsiCommand(ctx context.Context, path, content string) error {
	logf("write scsi file [path: %s content: %s]", path, content)
	args := []string{"-a", path}

	cmd := exec.CommandContext(ctx, BinaryTee, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, content)
	}()

	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command")
	}

	return nil
}
