package osbrick

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

// command binary
var (
	BinaryIscsiadm  = "iscsiadm"
	BinaryMultipath = "multipath"
	BinaryBlockdev  = "blockdev"
)

func iscsiadmBase(ctx context.Context, args []string) ([]byte, int, error) {
	logf("execute iscsiadm command [args: %s]", args)
	out, err := exec.CommandContext(ctx, BinaryIscsiadm, args...).CombinedOutput()
	if err != nil {
		e := err.(*exec.ExitError) // exec.Run() return ExitError and normal error, but this code not catch normal error.
		exitCode := e.Sys().(syscall.WaitStatus).ExitStatus()
		return nil, exitCode, fmt.Errorf("failed to execute command (args: %v): %w", args, err)
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
