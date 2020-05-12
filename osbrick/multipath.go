package osbrick

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
)

// ISCSIPath is connection of iSCSI volume
type ISCSIPath struct {
	PortalIP  string
	TargetIQN string
	HostLUNID int
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

		for i, ip := range ips {
			p := ISCSIPath{
				PortalIP:  ip,
				TargetIQN: iqns[i],
				HostLUNID: luns[i],
			}

			paths = append(paths, p)
		}
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
			logf("found dm device")
			break
		}

		logf("found err, continue... [device: %s] [err: %s]", d, err.Error())
		continue

	}

	return filepath.Join("/dev", dm), nil
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
