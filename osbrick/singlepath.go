package osbrick

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

// ConnectSinglePathVolume connect to iSCSI volume
func ConnectSinglePathVolume(ctx context.Context, targetPortalIP string, targetHostLUNID int) (string, error) {
	logf("Connecting volume (host lun ID: %d)", targetHostLUNID)

	ips, iqns, luns, err := GetIPsIQNsLUNs(ctx, targetPortalIP, targetHostLUNID)
	if err != nil {
		return "", fmt.Errorf("failed to get target info: %w", err)
	}
	paths := getiSCSIPath(ips, iqns, luns)
	if len(paths) != 1 {
		return "", fmt.Errorf("found multipath but call ConnectSinglePathVolume")
	}
	p := paths[0]

	device, err := connectVol(ctx, p.PortalIP, p.TargetIQN, p.HostLUNID)
	if err != nil {
		return "", fmt.Errorf("failed to connect volume: %w", err)
	}

	return filepath.Join("/dev", device), nil
}

// DisconnectSinglePathVolume disconnect single path volume
func DisconnectSinglePathVolume(ctx context.Context, targetPortalIP string, targetHostLUNID int) error {
	ips, iqns, luns, err := GetIPsIQNsLUNs(ctx, targetPortalIP, targetHostLUNID)
	if err != nil {
		return fmt.Errorf("failed to get target info: %w", err)
	}
	paths := getiSCSIPath(ips, iqns, luns)
	if len(paths) != 1 {
		return fmt.Errorf("found multipath but call DisconnectSinglePathVolume")
	}
	targetDevices, err := getConnectionDevices(ctx, paths)
	if err != nil {
		return fmt.Errorf("failed to get devices list: %w", err)
	}

	if err := removeConnection(ctx, targetDevices); err != nil {
		return fmt.Errorf("failed to remove connection: %w", err)
	}

	// check keep block device in same portal ip (from iscsiadm -m session -P3)
	attachedDevices, err := GetAttachedSCSIDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get attached devices: %w", err)
	}

	if len(attachedDevices) == 0 {
		// call logout when No action session
		if err := disconnectConnection(ctx, paths); err != nil {
			return fmt.Errorf("failed to disconnet iSCSI connection: %w", err)
		}
	}

	return nil
}

// removeSingleConnection remove iscsi session
// targetDeviceName example) "sdb"
func removeSingleConnection(ctx context.Context, targetDeviceName string) error {
	devicePath := filepath.Join("/dev", targetDeviceName)

	if err := removeScsiDevice(ctx, devicePath); err != nil {
		return fmt.Errorf("failed to remove iSCSI device: %w", err)
	}

	timeoutSecond := 10
	for i := 0; waitForVolumesRemoval([]string{targetDeviceName}); i++ {
		// until exist target volume.
		logf("wait removed target volume...")
		time.Sleep(1 * time.Second)

		if i == timeoutSecond {
			return fmt.Errorf("timeout exceeded wait for volume removal")
		}
	}

	if err := removeScsiSymlinks([]string{devicePath}); err != nil {
		return fmt.Errorf("failed to remove scsi symlinks: %w", err)
	}

	return nil
}