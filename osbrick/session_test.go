package osbrick

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"
)

func TestParseSessions(t *testing.T) {
	testInput := []string{
		"tcp: [1] 192.0.2.100:3260,1 iqn.0000-00.com.example:name1:192.0.2.100 (non-flash)", // single session
		`tcp: [3] 192.0.2.100:3260,1 iqn.0000-00.com.example:name2:192.0.2.100 (non-flash)
tcp: [5] 192.0.2.100:3260,2 iqn.0000-00.com.example:name100:192.0.2.100 (non-flash)`, // multi sessions
	}

	testOutput := [][]ISCSISession{
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
		sessions, err := ParseSessions([]byte(input))
		if err != nil {
			t.Error(err)
		}

		if reflect.DeepEqual(sessions, testOutput[i]) == false {
			t.Errorf("Unexpected a parsed result: %s", input)
		}
	}
}

const (
	testSessionP3 = `
iSCSI Transport Class version 2.0-870
version 2.0-874
Target: iqn.0000-00.com.example:name1:192.0.2.100 (non-flash)
        Current Portal: 192.0.2.100:3260,1
        Persistent Portal: 192.0.2.100:3260,1
                **********
                Interface:
                **********
                Iface Name: default
                Iface Transport: tcp
                Iface Initiatorname: iqn.0000-00.com.example:initiator0
                Iface IPaddress: 192.0.2.100:3260
                Iface HWaddress: <empty>
                Iface Netdev: <empty>
                SID: 189
                iSCSI Connection State: LOGGED IN
                iSCSI Session State: LOGGED_IN
                Internal iscsid Session State: NO CHANGE
                *********
                Timeouts:
                *********
                Recovery Timeout: 5
                Target Reset Timeout: 30
                LUN Reset Timeout: 30
                Abort Timeout: 15
                *****
                CHAP:
                *****
                username: <empty>
                password: ********
                username_in: <empty>
                password_in: ********
                ************************
                Negotiated iSCSI params:
                ************************
                HeaderDigest: None
                DataDigest: None
                MaxRecvDataSegmentLength: 262144
                MaxXmitDataSegmentLength: 262144
                FirstBurstLength: 73728
                MaxBurstLength: 262144
                ImmediateData: Yes
                InitialR2T: Yes
                MaxOutstandingR2T: 1
                ************************
                Attached SCSI devices:
                ************************
                Host Number: 2  State: running
                scsi2 Channel 00 Id 0 Lun: 0
                scsi2 Channel 00 Id 0 Lun: 1
                        Attached scsi disk sda          State: running
                scsi2 Channel 00 Id 0 Lun: 10
                        Attached scsi disk sdj          State: running
                scsi2 Channel 00 Id 0 Lun: 2
                        Attached scsi disk sdb          State: running
Target: iqn.0000-00.com.example:name2:192.0.2.100 (non-flash)
        Current Portal: 192.0.2.100:3260,1
        Persistent Portal: 192.0.2.100:3260,1
                **********
                Interface:
                **********
                Iface Name: default
                Iface Transport: tcp
                Iface Initiatorname: iqn.0000-00.com.example:initiator0
                Iface IPaddress: 192.0.2.100
                Iface HWaddress: <empty>
                Iface Netdev: <empty>
                SID: 190
                iSCSI Connection State: LOGGED IN
                iSCSI Session State: LOGGED_IN
                Internal iscsid Session State: NO CHANGE
                *********
                Timeouts:
                *********
                Recovery Timeout: 5
                Target Reset Timeout: 30
                LUN Reset Timeout: 30
                Abort Timeout: 15
                *****
                CHAP:
                *****
                username: <empty>
                password: ********
                username_in: <empty>
                password_in: ********
                ************************
                Negotiated iSCSI params:
                ************************
                HeaderDigest: None
                DataDigest: None
                MaxRecvDataSegmentLength: 262144
                MaxXmitDataSegmentLength: 262144
                FirstBurstLength: 4096
                MaxBurstLength: 262144
                ImmediateData: Yes
                InitialR2T: Yes
                MaxOutstandingR2T: 1
                ************************
                Attached SCSI devices:
                ************************
                Host Number: 3  State: running
                scsi3 Channel 00 Id 0 Lun: 0
                scsi3 Channel 00 Id 0 Lun: 1
                        Attached scsi disk sdk          State: running
                scsi3 Channel 00 Id 0 Lun: 10
                        Attached scsi disk sdt          State: running
                scsi3 Channel 00 Id 0 Lun: 2
                        Attached scsi disk sdl          State: running
`
)

func TestGetAttachedSCSIDevices(t *testing.T) {
	tests := []struct {
		input string
		want  []AttachedISCSIDevice
		err   bool
	}{
		{
			input: testSessionP3,
			want: []AttachedISCSIDevice{
				{
					TargetIQN:          "iqn.0000-00.com.example:name1:192.0.2.100",
					CurrentPortal:      "192.0.2.100:3260,1",
					HostID:             2,
					HostLUNID:          1,
					AttachedDeviceName: "sda",
				},
				{
					TargetIQN:          "iqn.0000-00.com.example:name1:192.0.2.100",
					CurrentPortal:      "192.0.2.100:3260,1",
					HostID:             2,
					HostLUNID:          10,
					AttachedDeviceName: "sdj",
				},
				{
					TargetIQN:          "iqn.0000-00.com.example:name1:192.0.2.100",
					CurrentPortal:      "192.0.2.100:3260,1",
					HostID:             2,
					HostLUNID:          2,
					AttachedDeviceName: "sdb",
				},
				{
					TargetIQN:          "iqn.0000-00.com.example:name2:192.0.2.100",
					CurrentPortal:      "192.0.2.100:3260,1",
					HostID:             3,
					HostLUNID:          1,
					AttachedDeviceName: "sdk",
				},
				{
					TargetIQN:          "iqn.0000-00.com.example:name2:192.0.2.100",
					CurrentPortal:      "192.0.2.100:3260,1",
					HostID:             3,
					HostLUNID:          10,
					AttachedDeviceName: "sdt",
				},
				{
					TargetIQN:          "iqn.0000-00.com.example:name2:192.0.2.100",
					CurrentPortal:      "192.0.2.100:3260,1",
					HostID:             3,
					HostLUNID:          2,
					AttachedDeviceName: "sdl",
				},
			},
			err: false,
		},
	}

	for _, test := range tests {
		targets, err := ParseSessionP3([]byte(test.input))
		if err != nil {
			t.Fatalf("failed to parse output: %v", err)
		}

		ds, err := getAttachedSCSIDevices(targets)
		if err != nil {
			t.Fatalf("failed to retrieve attached iSCSI device: %v", err)
		}

		if diff := deep.Equal(test.want, ds); diff != nil {
			t.Fatalf("want %q, but %q, diff %q:", test.want, ds, diff)
		}
	}

}
