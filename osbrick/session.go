package osbrick

import (
	"bufio"
	"bytes"
	"context"
	"errors"
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
		return nil, errors.New("invalid sentense")
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
