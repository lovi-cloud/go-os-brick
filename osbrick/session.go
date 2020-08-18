package osbrick

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ISCSISession is parsed information of session list
type ISCSISession struct {
	Transport            string
	SessionID            int
	TargetPortal         string
	TargetPortalGroupTag int
	IQN                  string
	NodeType             string
}

// GetSessions get parsed session list
func GetSessions(ctx context.Context) ([]ISCSISession, error) {
	out, err := getSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get output of iscsiadm -m session: %w", err)
	}

	return ParseSessions(out)
}

// getSessions get output of `iscsiadm -m session`
func getSessions(ctx context.Context) ([]byte, error) {
	out, exitCode, err := iscsiadmBase(ctx, []string{"-m", "session"})
	if err != nil {
		if exitCode == ExitCodeNoRecord {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to execute iscsiadm command: %w", err)
	}

	return out, nil
}

// ParseSessions parse output of iscsiadm -m session
func ParseSessions(out []byte) ([]ISCSISession, error) {
	var sessions []ISCSISession
	reader := bytes.NewReader(out)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		sentence := scanner.Text()
		s, err := parseSessionSentence(sentence)
		if err != nil {
			return nil, fmt.Errorf("failed to parse output sentense: %w", err)
		}

		sessions = append(sessions, *s)
	}

	return sessions, nil
}

func parseSessionSentence(sentence string) (*ISCSISession, error) {
	// sentence ex) transport_name: [session_id] ip_address:port,tpgt iqn node_type
	s := &ISCSISession{}

	info := strings.Split(sentence, " ")
	if len(info) != 5 {
		return nil, fmt.Errorf("invalid sentense (%s)", sentence)
	}

	s.Transport = strings.Trim(info[0], ":")

	sidStr := info[1][1 : len(info[1])-1]
	sid, err := strconv.Atoi(sidStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse session id: %w", err)
	}
	s.SessionID = sid

	iptp := strings.Split(info[2], ",")
	s.TargetPortal = iptp[0]
	tpgt, err := strconv.Atoi(iptp[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse target port group tag: %w", err)
	}
	s.TargetPortalGroupTag = tpgt

	s.IQN = info[3]
	s.NodeType = info[4]

	return s, nil
}

// AttachedISCSIDevice is device info
type AttachedISCSIDevice struct {
	TargetIQN          string
	CurrentPortal      string
	HostID             int
	HostLUNID          int
	AttachedDeviceName string // ex: sda, sdb
}

// GetAttachedSCSIDevices retrieves attached iSCSI devices.
func GetAttachedSCSIDevices(ctx context.Context) ([]AttachedISCSIDevice, error) {
	out, err := getSessionP3(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get output of iscsiadm -m session -P3: %w", err)
	}

	targets, err := ParseSessionP3(out)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output iscsiadm -m session -P3: %w", err)
	}

	devices, err := getAttachedSCSIDevices(targets)
	if err != nil {
		return nil, fmt.Errorf("failed to get attached device: %w", err)
	}

	return devices, nil
}

func getSessionP3(ctx context.Context) ([]byte, error) {
	out, exitCode, err := iscsiadmBase(ctx, []string{"-m", "session", "-P", "3"})
	if err != nil {
		if exitCode == ExitCodeNoRecord {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to execute iscsiadm command: %w", err)
	}

	return out, nil
}

// SessionP3Target is detail of iSCSI session per iSCSI target
type SessionP3Target struct {
	IQN              string
	CurrentPortal    string
	PersistentPortal string
	Blocks           []SessionP3Block
}

// SessionP3Block is content of iSCSI session
type SessionP3Block struct {
	Header string
	Body   []string
}

// ParseSessionP3 parse output of `iscsiadm -m session -P3
func ParseSessionP3(out []byte) ([]SessionP3Target, error) {
	reader := bytes.NewReader(out)
	scanner := bufio.NewScanner(reader)

	var targets []SessionP3Target
	var sentences []string

	for scanner.Scan() {
		sentence := scanner.Text()
		sentence = standardizeSpaces(sentence)

		if strings.HasPrefix(sentence, "Target:") && len(sentences) != 0 {
			// collected
			st, err := parseSessionP3Target(sentences)
			if err != nil {
				return nil, fmt.Errorf("failed to parse target block: %w", err)
			}
			targets = append(targets, st)
			sentences = []string{}
		}

		sentences = append(sentences, sentence)
	}

	if len(sentences) != 0 {
		st, err := parseSessionP3Target(sentences)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target last block: %w", err)
		}
		targets = append(targets, st)
	}

	return targets, nil
}

func parseSessionP3Target(target []string) (SessionP3Target, error) {
	var t SessionP3Target
	var block SessionP3Block
	var sentences []string

	for i, sentence := range target {
		if strings.HasPrefix(sentence, "Target:") {
			s := strings.Split(sentence, " ")
			if len(s) != 3 {
				return SessionP3Target{}, fmt.Errorf("invalid sentence (%s)", sentence)
			}

			t.IQN = s[1]
		}

		if strings.HasPrefix(sentence, "Current Portal") {
			s := strings.Split(sentence, " ")
			if len(s) != 3 {
				return SessionP3Target{}, fmt.Errorf("invalid sentence (%s)", sentence)
			}

			t.CurrentPortal = s[2]
		}

		if strings.HasPrefix(sentence, "Persistent Portal") {
			s := strings.Split(sentence, " ")
			if len(s) != 3 {
				return SessionP3Target{}, fmt.Errorf("invalid sentence (%s)", sentence)
			}

			t.CurrentPortal = s[2]
		}

		if strings.Contains(sentence, "*") && len(target[i:]) >= 3 && target[i] == target[i+2] {
			// header
			if block.Header != "" {
				// found next header and end of block
				block.Body = sentences
				t.Blocks = append(t.Blocks, block)

				sentences = []string{}
			}

			block.Header = strings.TrimSpace(target[i+1])
		}

		sentences = append(sentences, sentence)
	}

	if len(sentences) != 0 {
		block.Body = sentences
		t.Blocks = append(t.Blocks, block)
	}

	return t, nil
}

// getAttachedSCSIDevices retrieve attached devices from target
func getAttachedSCSIDevices(targets []SessionP3Target) ([]AttachedISCSIDevice, error) {
	var devices []AttachedISCSIDevice

	for _, target := range targets {
		for _, block := range target.Blocks {
			if block.Header == "Attached SCSI devices:" {
				ds, err := parseAttachedSCSIDevices(block.Body, target.IQN, target.CurrentPortal)
				if err != nil {
					return nil, fmt.Errorf("failed to parse attached devices: %s", err)
				}

				for _, d := range ds {
					if d.AttachedDeviceName != "" {
						// only session if AttachedDeviceName is blank
						devices = append(devices, d)
					}
				}

			}
		}
	}

	return devices, nil
}

func parseAttachedSCSIDevices(sentences []string, iqn, currentPortal string) ([]AttachedISCSIDevice, error) {
	var devices []AttachedISCSIDevice
	var hostID int

	for i, sentence := range sentences {
		if strings.HasPrefix(sentence, "Host Number") {
			s := strings.Split(sentence, " ")
			if len(s) != 5 {
				return nil, fmt.Errorf("invalid sentence, splited length is not 6 (%s)", sentence)
			}

			id, err := strconv.Atoi(s[2])
			if err != nil {
				return nil, fmt.Errorf("failed to convert host LUN ID: %w", err)
			}
			hostID = id
		}

		if strings.Contains(sentence, "Channel 00 Id") && len(sentences[i:]) >= 2 && strings.Contains(sentences[i+1], "Attached scsi disk") {
			// example)
			// scsi1 Channel 00 Id 0 Lun: 1
			//         Attached scsi disk sda         State: running
			//
			// NOTE(whywaita): what is 00?
			hostLUNID, err := parseHostLUNID(sentence)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve host LUN ID: %w", err)
			}

			s2 := strings.Split(sentences[i+1], " ")
			if len(s2) != 6 {
				return nil, fmt.Errorf("invalid sentence, splited length is not 6 (%s)", sentence)
			}

			d := AttachedISCSIDevice{
				TargetIQN:          iqn,
				CurrentPortal:      currentPortal,
				HostID:             hostID,
				HostLUNID:          hostLUNID,
				AttachedDeviceName: s2[3],
			}
			devices = append(devices, d)
		}

		if strings.Contains(sentence, "Channel 00 Id") && len(sentences[i:]) >= 2 && strings.Contains(sentences[i+1], "Channel 00 Id") {
			// example)
			// scsi1 Channel 00 Id 0 Lun: 0 <- parse only sentence
			// scsi1 Channel 00 Id 0 Lun: 1
			//
			// not attached disk
			hostLUNID, err := parseHostLUNID(sentence)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve host LUN ID: %w", err)
			}

			d := AttachedISCSIDevice{
				TargetIQN:          iqn,
				CurrentPortal:      currentPortal,
				HostID:             hostID,
				HostLUNID:          hostLUNID,
				AttachedDeviceName: "",
			}
			devices = append(devices, d)
		}
	}

	return devices, nil
}

func parseHostLUNID(sentence string) (int, error) {
	// example)
	// scsi1 Channel 00 Id 0 Lun: 0

	s := strings.Split(sentence, " ")
	if len(s) != 7 {
		return -1, fmt.Errorf("invalid sentence, splited length is not 7 (%s)", sentence)
	}

	hostLUNID, err := strconv.Atoi(s[6])
	if err != nil {
		return -1, fmt.Errorf("failed to convert host LUN ID: %w", err)
	}

	return hostLUNID, nil
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
