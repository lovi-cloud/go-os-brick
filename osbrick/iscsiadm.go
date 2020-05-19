package osbrick

import (
	"context"
	"fmt"
)

func iscsiadm(ctx context.Context, portalIP, iqn string, args []string) ([]byte, int, error) {
	var a []string
	baseArgs := []string{"-m", "node"}
	a = append(baseArgs, []string{"-T", iqn}...)
	a = append(a, []string{"-p", portalIP}...)
	a = append(a, args...)

	out, exitCode, err := iscsiadmBase(ctx, a)
	if err != nil {
		return nil, exitCode, fmt.Errorf("failed to execute iscsiadm command: %w", err)
	}

	return out, exitCode, nil
}

func iscsiadmUpdate(ctx context.Context, portalIP, targetIQN, key, value string, args []string) ([]byte, int, error) {
	a := []string{"--op", "update", "-n", key, "-v", value}
	a = append(a, args...)
	return iscsiadm(ctx, portalIP, targetIQN, a)
}
