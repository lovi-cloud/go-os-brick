package osbrick

import (
	"context"
	"errors"
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

func connectVol(ctx context.Context, portalIP, targetIqn string, targetHostLunID int) (string, error) {
	sessionID, err := connectToiSCSIPortal(ctx, portalIP, targetIqn, 0)
	if err != nil {
		return "", fmt.Errorf("failed to connect iSCSI portal: %s", err)
	}

	hctl, err := GetHctl(sessionID, targetHostLunID)
	if err != nil {
		return "", fmt.Errorf("failed to get hctl: %w", err)
	}

	if err := scanISCSI(hctl); err != nil {
		return "", fmt.Errorf("failed to rescan target: %w", err)
	}

	device, err := GetDeviceName(sessionID, hctl)
	if err != nil {
		return "", fmt.Errorf("failed to get device name: %w", err)
	}

	logf("connected to %s", device)
	return device, nil
}

// connectPortal connect to iSCSI Portal via target IQN.
// return session id.
func connectToiSCSIPortal(ctx context.Context, portalIP, targetIQN string, retryCount int) (int, error) {
	if retryCount == 0 {
		retryCount = 10
	}

	//// must be node.session.scan is manual
	//_, _, err := iscsiadmUpdate(ctx, portalIP, targetIQN, "node.session.scan", "manual", nil)
	//if err != nil {
	//	return 0, fmt.Errorf("failed to update node.session.scan to manual: %w", err)
	//}

	// NOTE(whywaita): add while loop if issue of find session
	if err := LoginPortal(ctx, portalIP, targetIQN); err != nil {
		return 0, fmt.Errorf("failed to iSCSI portal login: %w", err)
	}
	for i := 0; i < retryCount; i++ {
		sessions, err := GetSessions(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get iSCSI sessions: %w", err)
		}

		for _, session := range sessions {
			if session.TargetPortal == portalIP && session.IQN == targetIQN {
				// found session
				return session.SessionID, nil
			}
		}

		time.Sleep(1 * time.Second)
	}

	return -1, errors.New("session id is not found")
}
