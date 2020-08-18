package osbrick

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

func iscsiadmBase(ctx context.Context, args []string) ([]byte, int, error) {
	logf("execute iscsiadm command [args: %s]", args)
	out, err := exec.CommandContext(ctx, BinaryIscsiadm, args...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// it's ExitError
			exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
			return nil, exitCode, fmt.Errorf("failed to execute command (args: %v): %w", args, err)
		}

		return nil, 1, fmt.Errorf("failed to execute command (args: %v): %w", args, err)
	}

	return out, 0, nil
}

func multipathBase(ctx context.Context, args []string) ([]byte, int, error) {
	logf("execute multipath command [args: %s]", args)
	out, err := exec.CommandContext(ctx, BinaryMultipath, args...).CombinedOutput()
	if err != nil {
		e := err.(*exec.ExitError) // exec.Run() return ExitError and normal error, but this code not catch normal error.
		exitCode := e.Sys().(syscall.WaitStatus).ExitStatus()
		return nil, exitCode, fmt.Errorf("failed to execute command (args: %v): %w", args, err)
	}

	return out, 0, nil
}
