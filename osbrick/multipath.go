package osbrick

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ISCSIPath is connection of iSCSI volume
type ISCSIPath struct {
	PortalIP  string
	TargetIQN string
	HostLUNID int
}

func getiSCSIPath(ips, iqns []string, luns []int) []ISCSIPath {
	var paths []ISCSIPath

	for i, ip := range ips {
		p := ISCSIPath{
			PortalIP:  ip,
			TargetIQN: iqns[i],
			HostLUNID: luns[i],
		}

		paths = append(paths, p)
	}

	return paths
}

// ConnectMultipathVolume connect to iSCSI volume using multipath.
func ConnectMultipathVolume(ctx context.Context, targetPortalIPs []string, targetHostLUNID int) (string, error) {
	var paths []ISCSIPath
	var err error
	for _, portalIP := range targetPortalIPs {
		ips, iqns, luns, err := GetIPsIQNsLUNs(ctx, portalIP, targetHostLUNID)
		if err != nil {
			return "", fmt.Errorf("failed to get target info: %w", err)
		}

		ps := getiSCSIPath(ips, iqns, luns)
		paths = append(paths, ps...)
	}

	var wg sync.WaitGroup
	var devices []string
	for _, p := range paths {
		wg.Add(1)
		device, err := connectVol(ctx, p.PortalIP, p.TargetIQN, p.HostLUNID)
		if err != nil {
			return "", fmt.Errorf("failed to connect volume: %w", err)
		}
		devices = append(devices, device)
		wg.Done()
	}
	wg.Wait()
	// NOTE(whywaita): will implement using goroutine channel if occurred performance issue.

	var dm string
	for _, d := range devices {
		dm, err = findSysfsMultipathDM(d)
		if err == nil {
			logf("found dm device: %v", dm)
			break
		}

		logf("found err, continue... [device: %s] [err: %s]", d, err.Error())
		continue
	}

	return filepath.Join("/dev", dm), nil
}

// DisconnectVolume disconnect volume
func DisconnectVolume(ctx context.Context, targetPortalIPs []string, targetHostLUNID int) error {
	return cleanupConnection(ctx, targetPortalIPs, targetHostLUNID)
}

func cleanupConnection(ctx context.Context, targetPortalIPs []string, targetHostLUNID int) error {
	var paths []ISCSIPath
	for _, portalIP := range targetPortalIPs {
		ips, iqns, luns, err := GetIPsIQNsLUNs(ctx, portalIP, targetHostLUNID)
		if err != nil {
			return fmt.Errorf("failed to get ips, iqns, luns: %w", err)
		}
		ps := getiSCSIPath(ips, iqns, luns)
		paths = append(paths, ps...)
	}

	targetDevices, err := getConnectionDevices(ctx, paths)
	if err != nil {
		return fmt.Errorf("failed to get devices list: %w", err)
	}

	err = removeConnection(ctx, targetDevices)
	if err != nil {
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

// disconnectConnection logout from portal ip.
// this function delete all device from p.PortalIP
func disconnectConnection(ctx context.Context, paths []ISCSIPath) error {
	for _, p := range paths {
		err := disconnectFromIscsiPortal(ctx, p.PortalIP, p.TargetIQN)
		if err != nil {
			return fmt.Errorf("failed to disconnect from iSCSI portal: %w", err)
		}
	}

	return nil
}

func disconnectFromIscsiPortal(ctx context.Context, portalIP, targetIQN string) error {
	_, _, err := iscsiadmUpdate(ctx, portalIP, targetIQN, "node.startup", "manual", nil)
	if err != nil {
		return fmt.Errorf("failed to update node.startup to manual: %w", err)
	}

	err = LogoutPortal(ctx, portalIP, targetIQN)
	if err != nil {
		return fmt.Errorf("failed to logout portal: %w", err)
	}

	_, _, err = iscsiadm(ctx, portalIP, targetIQN, []string{"--op", "delete"})
	if err != nil {
		return fmt.Errorf("failed to execute --op delete: %w", err)
	}

	return nil
}

// removeConnection remove iscsi multipath session
// targetDeviceNames example) []string{"sda", "sdb"}
func removeConnection(ctx context.Context, targetDeviceNames []string) error {
	if targetDeviceNames == nil {
		return errors.New("targetDeviceNames is nil")
	}
	var devicePaths []string
	for _, dn := range targetDeviceNames {
		devicePaths = append(devicePaths, "/dev/"+dn)
	}

	multipathDeviceName, err := findSysfsMultipathDM(targetDeviceNames[0]) // targetDeviceNames have a same dm device.
	if err != nil {
		return fmt.Errorf("failed to find multipath device volume name: %w", err)
	}
	multipathDevicePath := "/dev/" + multipathDeviceName

	err = flushMultipathDevice(ctx, multipathDevicePath)
	if err != nil {
		return fmt.Errorf("failed to flush multipath device")
	}

	for _, devicePath := range devicePaths {
		err := removeScsiDevice(ctx, devicePath)
		if err != nil {
			return fmt.Errorf("failed to remove iSCSI device: %w", err)
		}
	}

	timeoutSecond := 10
	for i := 0; waitForVolumesRemoval(targetDeviceNames); i++ {
		// until exist target volume.
		logf("wait removed target volume...")
		time.Sleep(1 * time.Second)

		if i == timeoutSecond {
			return fmt.Errorf("timeout exceeded wait for volume removal")
		}
	}

	err = removeScsiSymlinks(devicePaths)
	if err != nil {
		return fmt.Errorf("failed to remove scsi symlinks: %w", err)
	}

	return nil
}

func removeScsiDevice(ctx context.Context, devicePath string) error {
	deviceName := strings.TrimPrefix(devicePath, "/dev/")
	deletePath := fmt.Sprintf("/sys/block/%s/device/delete", deviceName)
	_, err := os.Stat(deletePath)
	if err != nil {
		return fmt.Errorf("failed to stat device delete path: %w", err)
	}

	err = flushDeviceIO(ctx, devicePath)
	if err != nil {
		return fmt.Errorf("failed to flush device I/O: %w", err)
	}

	err = echoScsiCommand(ctx, deletePath, "1")
	if err != nil {
		return fmt.Errorf("failed to write to delete path: %w", err)
	}

	return nil
}

func flushDeviceIO(ctx context.Context, devicePath string) error {
	_, err := os.Stat(devicePath)
	if err != nil {
		return fmt.Errorf("failed to stat device path: %w", err)
	}

	_, err = exec.CommandContext(ctx, BinaryBlockdev, "--flushbufs", devicePath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute blockdev command: %s", err)
	}

	return nil
}

func flushMultipathDevice(ctx context.Context, targetMultipathPath string) error {
	_, _, err := multipathBase(ctx, []string{"-f", targetMultipathPath})
	if err != nil {
		return fmt.Errorf("failed to execute multipath device flush command: %w", err)
	}

	return nil
}

// waitForVolumesRemoval check target device. return true if until exist.
func waitForVolumesRemoval(targetDevicePaths []string) bool {
	exist := false

	for _, devicePath := range targetDevicePaths {
		_, err := os.Stat(devicePath)
		if err == nil {
			logf("found not deleted volume: %s", devicePath)
			exist = true
			break
		}
	}

	return exist
}

func removeScsiSymlinks(devicePaths []string) error {
	links, err := filepath.Glob("/dev/disk/by-id/scsi-*")
	if err != nil {
		return fmt.Errorf("failed to get scsi link")
	}

	var removeTarget []string
	for _, link := range links {
		realpath, err := filepath.EvalSymlinks(link)
		if err != nil {
			logf("failed to get realpath: %v", err)
		}

		for _, devicePath := range devicePaths {
			if realpath == devicePath {
				removeTarget = append(removeTarget, link)
				break
			}
		}
	}

	for _, l := range removeTarget {
		err = os.Remove(l)
		if err != nil {
			return fmt.Errorf("failed to delete symlink: %w", err)
		}
	}

	return nil
}

// getConnectionDevices get volumes in paths
// return device is only device name (ex: sda, sdb)
func getConnectionDevices(ctx context.Context, paths []ISCSIPath) ([]string, error) {
	var devices []string

	sessions, err := GetSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get iSCSI sessions: %w", err)
	}

	for _, path := range paths {
		for _, session := range sessions {
			if session.TargetPortal != path.PortalIP || session.IQN != path.TargetIQN {
				continue
			}

			hctl, err := GetHctl(session.SessionID, path.HostLUNID)
			if err != nil {
				return nil, fmt.Errorf("failed to get hctl info: %w", err)
			}
			deviceName, err := GetDeviceName(session.SessionID, hctl)
			if err != nil {
				return nil, fmt.Errorf("failed to get device name: %w", err)
			}
			if hctl.HostLUNID == path.HostLUNID {
				devices = append(devices, deviceName)
			}
		}
	}

	return devices, nil
}

func findSysfsMultipathDM(deviceName string) (dmDeviceName string, err error) {
	globStr := fmt.Sprintf("/sys/block/%s/holders/dm-*", deviceName)
	paths, err := filepath.Glob(globStr)
	if err != nil {
		return "", fmt.Errorf("failed to glob dm device filepath: %w", err)
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("dm device is not found")
	}

	_, name := filepath.Split(paths[0])
	return name, nil
}
