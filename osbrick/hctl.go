package osbrick

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Hctl is IDs of SCSI
type Hctl struct {
	HostID    int
	ChannelID int
	TargetID  int
	HostLUNID int
}

// GetHctl search a some ID by given session ID.
// return ID of host, channel, target.
func GetHctl(sessionID, hostLUNID int) (*Hctl, error) {
	globStr := fmt.Sprintf("/sys/class/iscsi_host/host*/device/session%d/target*", sessionID)
	paths, err := filepath.Glob(globStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions filepath: %w", err)
	}
	if len(paths) != 1 {
		return nil, fmt.Errorf("target filepath is not found")
	}

	_, filename := filepath.Split(paths[0])
	ids := strings.Split(filename, ":") // ex: target1:0:0
	if len(ids) != 3 {
		return nil, fmt.Errorf("failed to parse iSCSI session filename")
	}
	channelID, err := strconv.Atoi(ids[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse channel ID: %w", err)
	}
	targetID, err := strconv.Atoi(ids[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse target ID: %w", err)
	}

	names := strings.Split(paths[0], "/")
	hostIDstr := strings.TrimPrefix(searchHost(names), "host")
	hostID, err := strconv.Atoi(hostIDstr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host ID: %w", err)
	}

	hctl := &Hctl{
		HostID:    hostID,
		ChannelID: channelID,
		TargetID:  targetID,
		HostLUNID: hostLUNID,
	}

	return hctl, nil
}

// searchHost search param
// return "host"+id
func searchHost(names []string) string {
	for _, v := range names {
		if strings.HasPrefix(v, "host") {
			return v
		}
	}

	return ""
}

// GetDeviceName get device name of connected volume
func GetDeviceName(sessionID int, hctl *Hctl) (string, error) {
	var lastErr error

	for i := 0; i < 5*60; i++ {
		// retry 5 minutes
		deviceName, err := getDeviceName(sessionID, hctl)
		if err == nil {
			return deviceName, nil
		}

		logf("failed to get device name, do retry: %+v", err)
		lastErr = err
		time.Sleep(1 * time.Second)
	}

	return "", lastErr
}

func getDeviceName(sessionID int, hctl *Hctl) (string, error) {
	p := fmt.Sprintf(
		"/sys/class/iscsi_host/host%d/device/session%d/target%d:%d:%d/%d:%d:%d:%d/block/*",
		hctl.HostID,
		sessionID,
		hctl.HostID, hctl.ChannelID, hctl.TargetID,
		hctl.HostID, hctl.ChannelID, hctl.TargetID, hctl.HostLUNID)

	paths, err := filepath.Glob(p)
	if err != nil {
		return "", fmt.Errorf("failed to parse iSCSI block device filepath: %w", err)
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("device filepath is not found")
	}

	_, deviceName := filepath.Split(paths[0])

	return deviceName, nil
}
