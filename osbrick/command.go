package osbrick

import (
	"context"
	"fmt"
	"os"
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
			return nil, exitCode, fmt.Errorf("failed to execute command (args: %v, out: %s): %w", args, string(out), err)
		}

		return nil, 1, fmt.Errorf("failed to execute command (args: %v, out: %s): %w", args, string(out), err)
	}

	return out, 0, nil
}

func multipathBase(ctx context.Context, args []string) ([]byte, int, error) {
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

func echoScsiCommand(path, content string) error {
	// write content to path (sysfs)
	logf("write scsi file [path: %s content: %s]", path, content)

	f, err := os.OpenFile(path, os.O_WRONLY, 0400)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func qemuimgConvertBase(ctx context.Context, args []string) error {
	c := []string{"convert"}
	a := append(c, args...)

	logf("execute qemu-img command [args: %s]", a)
	out, err := exec.CommandContext(ctx, "qemu-img", a...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute qemu-img convert command (args: %s, out: %s): %w", args, string(out), err)
	}

	return nil
}

// QEMUToRaw convert os image.
func QEMUToRaw(ctx context.Context, src, dest string) error {
	args := []string{"-O", "raw", "-t", "none", "-f", "qcow2", src, dest}

	if err := qemuimgConvertBase(ctx, args); err != nil {
		return fmt.Errorf("failed to execute convert command: %w", err)
	}

	return nil
}
