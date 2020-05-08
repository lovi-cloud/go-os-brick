package osbrick

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func connectVol(ctx context.Context, portalIP, targetIqn, targetHostLunID string) (string, error) {
	sessionID, err := connectToiSCSIPortal(ctx, portalIP, targetIqn, 0)
	if err != nil {
		return "", fmt.Errorf("failed to connect iSCSI portal: %s", err)
	}

	hctl, err := GetHctl(sessionID, targetHostLunID)
	if err != nil {
		return "", fmt.Errorf("failed to get hctl: %w", err)
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
func connectToiSCSIPortal(ctx context.Context, portalIP, targetIqn string, retryCount int) (int, error) {
	if retryCount == 0 {
		retryCount = 10
	}

	err := LoginPortal(ctx, portalIP, targetIqn)
	if err != nil {
		return 0, fmt.Errorf("failed to iSCSI portal login: %w", err)
	}
	for i := 0; i < retryCount; i++ {
		sessions, err := GetSessions(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get iSCSI sessions: %w", err)
		}

		for _, session := range sessions {
			if session.TargetPortal == portalIP && session.IQN == targetIqn {
				// found session
				return session.SessionID, nil
			}
		}

		time.Sleep(1 * time.Second)
	}

	return -1, errors.New("session id is not found")
}
