package osbrick

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

func iscsiadmBase(ctx context.Context, args []string) ([]byte, int, error) {
	commandMu.Lock()
	defer commandMu.Unlock()

	logf("execute iscsiadm command [args: %s]", args)
	out, err := exec.CommandContext(ctx, BinaryIscsiadm, args...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// it's ExitError
			exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
			return nil, exitCode, fmt.Errorf("failed to execute command (args: %v, out: %s): %w", args, string(out), err)
		}

		return nil, 1, fmt.Errorf("failed to execute command (args: %v, out: %s): %w", args, string(out), err)
	}

	return out, 0, nil
}

func multipathBase(ctx context.Context, args []string) ([]byte, int, error) {
	commandMu.Lock()
	defer commandMu.Unlock()

	logf("execute multipath command [args: %s]", args)
	out, err := exec.CommandContext(ctx, BinaryMultipath, args...).CombinedOutput()
	if err != nil {
		e := err.(*exec.ExitError) // exec.Run() return ExitError and normal error, but this code not catch normal error.
		exitCode := e.Sys().(syscall.WaitStatus).ExitStatus()
		return nil, exitCode, fmt.Errorf("failed to execute command (args: %v, out: %s): %w", args, string(out), err)
	}

	return out, 0, nil
}

func blockdevBase(ctx context.Context, args []string) ([]byte, int, error) {
	commandMu.Lock()
	defer commandMu.Unlock()

	logf("execute blockdev command [args: %s]", args)
	out, err := exec.CommandContext(ctx, BinaryBlockdev, args...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// it's ExitError
			exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
			return nil, exitCode, fmt.Errorf("failed to execute command (args: %v, out: %s): %w", args, string(out), err)
		}

		return nil, 1, fmt.Errorf("failed to execute command (args: %v, out: %s): %w", args, string(out), err)
	}

	return out, 0, nil
}

func echoScsiCommand(ctx context.Context, path, content string) error {
	commandMu.Lock()
	defer commandMu.Unlock()

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

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command (out: %s): %w", string(out), err)
	}

	return nil
}
