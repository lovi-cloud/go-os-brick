package osbrick

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

func iscsiadmBase(ctx context.Context, args []string) ([]byte, int, error) {
	logf("execute iscsiadm command [args: %s]", args)
	out, err := exec.CommandContext(ctx, "iscsiadm", args...).CombinedOutput()
	if err != nil {
		e := err.(*exec.ExitError) // exec.Run() return ExitError and normal error, but this code not catch normal error.
		exitCode := e.Sys().(syscall.WaitStatus).ExitStatus()
		return nil, exitCode, fmt.Errorf("failed to execute command (args: %v): %w", args, err)
	}

	return out, 0, nil
}

func iscsiadm(ctx context.Context, iqn, portalIP string, args []string) ([]byte, int, error) {
	var a []string
	baseArgs := []string{"-m", "node"}
	if iqn != "" {
		a = append(baseArgs, []string{"-t", iqn}...)
	}
	a = append(a, []string{"-p", portalIP}...)
	a = append(baseArgs, args...)

	out, exitCode, err := iscsiadmBase(ctx, a)
	if err != nil {
		return nil, exitCode, fmt.Errorf("failed to execute iscsiadm command (args: %s): %w", args, err)
	}

	return out, exitCode, nil
}

func iscsiadmUpdate(ctx context.Context, portalIP, key, value string, args []string) ([]byte, int, error) {
	a := []string{"--op", "update", "-n", key, "-v", value}
	a = append(a, args...)
	return iscsiadm(ctx, "", portalIP, a)
}
