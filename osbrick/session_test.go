package osbrick_test

import (
	"reflect"
	"testing"

	"github.com/whywaita/go-os-brick/osbrick"
)

func TestParseSessions(t *testing.T) {
	testInput := []string{
		"tcp: [1] 192.0.2.100:3260,1 iqn.0000-00.com.example:name1:192.0.2.100 (non-flash)", // single session
		`tcp: [3] 192.0.2.100:3260,1 iqn.0000-00.com.example:name2:192.0.2.100 (non-flash)
tcp: [5] 192.0.2.100:3260,2 iqn.0000-00.com.example:name100:192.0.2.100 (non-flash)`, // multi sessions
	}

	testOutput := [][]osbrick.ISCSISession{
		{
			{
				Transport:            "tcp",
				SessionID:            1,
				TargetPortal:         "192.0.2.100:3260",
				TargetPortalGroupTag: 1,
				IQN:                  "iqn.0000-00.com.example:name1:192.0.2.100",
				NodeType:             "(non-flash)",
			},
		},
		{
			{
				Transport:            "tcp",
				SessionID:            3,
				TargetPortal:         "192.0.2.100:3260",
				TargetPortalGroupTag: 1,
				IQN:                  "iqn.0000-00.com.example:name2:192.0.2.100",
				NodeType:             "(non-flash)",
			},
			{
				Transport:            "tcp",
				SessionID:            5,
				TargetPortal:         "192.0.2.100:3260",
				TargetPortalGroupTag: 2,
				IQN:                  "iqn.0000-00.com.example:name100:192.0.2.100",
				NodeType:             "(non-flash)",
			},
		},
	}

	for i, input := range testInput {
		sessions, err := osbrick.ParseSessions([]byte(input))
		if err != nil {
			t.Error(err)
		}

		if reflect.DeepEqual(sessions, testOutput[i]) == false {
			t.Errorf("Unexpected a parsed result: %s", input)
		}
	}
}
